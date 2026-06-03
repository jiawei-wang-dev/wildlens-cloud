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