import boto3
import json
import hashlib
import requests
from datetime import datetime

# Configuration
GCP_INFER_URL = "PLACEHOLDER_REPLACE_WITH_GCP_URL"
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


def write_to_dynamodb(file_id, bucket, key, file_type, tags, tag_counts, primary_species, model_version, thumbnail_path):
    """Write media metadata to DynamoDB"""
    table = dynamodb.Table(DYNAMODB_TABLE)
    
    file_url = f"https://{bucket}.s3.amazonaws.com/{key}"
    thumbnail_url = f"https://{bucket}.s3.amazonaws.com/{thumbnail_path}" if thumbnail_path else None
    
    item = {
        'file_id': file_id,
        'file_url': file_url,
        'thumbnail_url': thumbnail_url,
        'file_type': file_type,
        'tags': tags,
        'tag_counts': tag_counts,
        'primary_species': primary_species,
        'model_version': model_version,
        'status': 'ready',
        'created_at': datetime.utcnow().isoformat()
    }
    
    table.put_item(Item=item)
    print(f"Written to DynamoDB: {file_id}")
    return file_url