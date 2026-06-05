import boto3
import json
import hashlib
from datetime import datetime

# Configuration
S3_BUCKET = "fit5225-wildlife-media"
s3_client = boto3.client('s3', region_name='us-east-1')

def generate_presigned_upload_url(object_path, mime_type, expiry=300):
    """Generate a presigned URL for uploading a file to S3"""
    url = s3_client.generate_presigned_url(
        'put_object',
        Params={
            'Bucket': S3_BUCKET,
            'Key': object_path,
            'ContentType': mime_type
        },
        ExpiresIn=expiry
    )
    return url

def lambda_handler(event, context):
    """
    Generate a presigned S3 upload URL for the frontend
    
    Expected request body:
    {
        "filename": "koala.jpg",
        "checksum_sha256": "abc123...",
        "file_type": "image",
        "mime_type": "image/jpeg",
        "owner_id": "user@email.com"
    }
    """
    try:
        # Parse request body
        body = json.loads(event.get('body', '{}'))
        
        filename = body.get('filename')
        checksum = body.get('checksum_sha256')
        mime_type = body.get('mime_type', 'image/jpeg')
        owner_id = body.get('owner_id', 'anonymous')
        
        if not filename or not checksum:
            return {
                'statusCode': 400,
                'headers': {'Access-Control-Allow-Origin': '*'},
                'body': json.dumps({'error': 'filename and checksum_sha256 are required'})
            }
        
        # Check for duplicate
        dynamodb = boto3.resource('dynamodb', region_name='us-east-1')
        table = dynamodb.Table('fit5225-wildlife-media-metadata')
        response = table.get_item(Key={'file_id': checksum})
        if 'Item' in response:
            return {
                'statusCode': 200,
                'headers': {'Access-Control-Allow-Origin': '*'},
                'body': json.dumps({
                    'duplicate': True,
                    'file_id': checksum,
                    'message': 'File already exists'
                })
            }
        
        # Build object path as per B's contract
        object_path = f"incoming/{owner_id}/{checksum}/{filename}"
        
        # Generate presigned upload URL
        upload_url = generate_presigned_upload_url(object_path, mime_type)
        
        return {
            'statusCode': 200,
            'headers': {'Access-Control-Allow-Origin': '*'},
            'body': json.dumps({
                'upload_url': upload_url,
                'object_path': object_path,
                'bucket': S3_BUCKET,
                'file_id': checksum
            })
        }
    
    except Exception as e:
        return {
            'statusCode': 500,
            'headers': {'Access-Control-Allow-Origin': '*'},
            'body': json.dumps({'error': str(e)})
        }