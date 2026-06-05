import boto3
import json
import hashlib
import requests
from datetime import datetime

# Configuration
GCP_INFER_URL = "https://wildlens-media-infer-343888474330.australia-southeast1.run.app/infer"
DYNAMODB_TABLE = "fit5225-wildlife-media-metadata"
SNS_NOTIFICATION_FUNCTION = "wildlife-sns-notification"

# AWS clients
s3_client = boto3.client('s3', region_name='us-east-1')
dynamodb = boto3.resource('dynamodb', region_name='us-east-1')
lambda_client = boto3.client('lambda', region_name='us-east-1')

def get_s3_presigned_url(bucket, key, expiry=300):
    """Generate a short-lived presigned URL for S3 object"""
    url = s3_client.generate_presigned_url(
        'get_object',
        Params={'Bucket': bucket, 'Key': key},
        ExpiresIn=expiry
    )
    return url

def check_duplicate(file_id):
    """Check if file already exists in DynamoDB using checksum"""
    table = dynamodb.Table(DYNAMODB_TABLE)
    response = table.get_item(Key={'file_id': file_id})
    return 'Item' in response


def write_to_dynamodb(file_id, bucket, key, file_type, filename, mime_type, tags, tag_counts, primary_species, model_version, thumbnail_path):
    """Write media metadata to DynamoDB"""
    table = dynamodb.Table(DYNAMODB_TABLE)
    
    file_url = f"https://{bucket}.s3.amazonaws.com/{key}"
    thumbnail_url = f"https://{bucket}.s3.amazonaws.com/{thumbnail_path}" if thumbnail_path else None
    now = datetime.utcnow().isoformat()
    
    item = {
        'file_id': file_id,
        'checksum_sha256': file_id,
        'bucket': bucket,
        'object_path': key,
        'original_filename': filename,
        'mime_type': mime_type,
        'file_type': file_type,
        'file_url': file_url,
        'thumbnail_url': thumbnail_url,
        'thumbnail_object_path': thumbnail_path,
        'tags': tags,
        'tag_counts': tag_counts,
        'primary_species': primary_species,
        'model_version': model_version,
        'status': 'ready',
        'storage_provider': 'aws',
        'created_at': now,
        'updated_at': now
    }
    
    table.put_item(Item=item)
    print(f"Written to DynamoDB: {file_id}")
    return file_url

def call_gcp_infer(bucket, key, file_id, file_type, presigned_url=None):
    """Call GCP Cloud Run ML inference service"""
    if GCP_INFER_URL == "PLACEHOLDER_REPLACE_WITH_GCP_URL":
        print("WARNING: GCP inference URL not configured yet")
        return None
    
    filename = key.split('/')[-1]
    
    # Determine mime type
    if file_type == 'video':
        mime_type = 'video/mp4'
    else:
        mime_type = 'image/jpeg'
    
    # Generate thumbnail upload URL
    thumbnail_object_path = f"media/thumbnails/{file_id}.jpg"
    thumbnail_upload_url = s3_client.generate_presigned_url(
        'put_object',
        Params={
            'Bucket': bucket,
            'Key': thumbnail_object_path,
            'ContentType': 'image/jpeg'
        },
        ExpiresIn=300
    )
    
    payload = {
        "file_id": file_id,
        "bucket": bucket,
        "object_path": key,
        "filename": filename,
        "file_type": file_type,
        "mime_type": mime_type,
        "checksum_sha256": file_id,
        "download_url": presigned_url,
        "thumbnail_upload_url": thumbnail_upload_url
    }
    
    headers = {
        "Content-Type": "application/json"
    }
    
    print(f"Sending to GCP: payload with thumbnail_upload_url")
    
    response = requests.post(
        GCP_INFER_URL,
        json=payload,
        headers=headers,
        timeout=60
    )
    
    print(f"GCP response status: {response.status_code}")
    print(f"GCP response body: {response.text}")
    
    if response.status_code == 200:
        result = response.json()
        return result
    else:
        print(f"GCP inference failed: {response.status_code}")
        return None
    
def trigger_sns_notification(tag_name, file_url, file_type):
    """Trigger SNS notification Lambda for each detected tag"""
    for tag in tag_name:
        payload = {
            "action": "notify_upload",
            "tag_name": tag,
            "file_url": file_url,
            "file_type": file_type
        }
        
        lambda_client.invoke(
            FunctionName=SNS_NOTIFICATION_FUNCTION,
            InvocationType='Event',
            Payload=json.dumps(payload)
        )
        print(f"SNS notification triggered for tag: {tag}")

def lambda_handler(event, context):
    """
    Lambda entry point - triggered by S3 upload event
    """
    for record in event['Records']:
        bucket = record['s3']['bucket']['name']
        key = record['s3']['object']['key']
        
        print(f"Processing file: s3://{bucket}/{key}")
        
        # Generate file_id from actual file checksum
        response = s3_client.get_object(Bucket=bucket, Key=key)
        file_content = response['Body'].read()
        file_id = hashlib.sha256(file_content).hexdigest()
                
        # Check for duplicate
        if check_duplicate(file_id):
            print(f"Duplicate file detected: {file_id}, skipping...")
            continue
        
        # Determine file type
        file_type = 'video' if key.lower().endswith(('.mp4', '.avi', '.mov')) else 'image'
        
        # Generate presigned URL for GCP
        presigned_url = get_s3_presigned_url(bucket, key)
        
        # Call GCP Cloud Run for ML inference
        infer_result = call_gcp_infer(
            bucket=bucket,
            key=key,
            file_id=file_id,
            file_type=file_type,
            presigned_url=presigned_url
        )
        
        if infer_result:
            # Write results to DynamoDB
            file_url = write_to_dynamodb(
                file_id=file_id,
                bucket=bucket,
                key=key,
                file_type=file_type,
                filename=key.split('/')[-1],
                mime_type='video/mp4' if file_type == 'video' else 'image/jpeg',
                tags=infer_result.get('tags', []),
                tag_counts=infer_result.get('tag_counts', {}),
                primary_species=infer_result.get('primary_species', ''),
                model_version=infer_result.get('model_version', ''),
                thumbnail_path=infer_result.get('thumbnail_object_path', '')
            )
            
            # Trigger SNS notifications
            trigger_sns_notification(
                infer_result.get('tags', []),
                file_url,
                file_type
            )
        else:
            print(f"Inference failed for {file_id}, skipping DynamoDB write")
    
    return {'statusCode': 200, 'body': 'Processing complete'}