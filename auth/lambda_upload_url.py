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