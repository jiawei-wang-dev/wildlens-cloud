import boto3
import json

# Initialize SNS client
sns_client = boto3.client('sns', region_name='us-east-1')

def create_topic_for_tag(tag_name):
    """Create an SNS topic for a specific wildlife tag if it doesn't exist"""
    topic_name = f"wildlife-tag-{tag_name.lower().replace(' ', '-')}"
    
    response = sns_client.create_topic(Name=topic_name)
    topic_arn = response['TopicArn']
    
    print(f"Topic created/found for tag '{tag_name}': {topic_arn}")
    return topic_arn


def subscribe_email_to_tag(email, tag_name):
    """Subscribe an email address to a specific tag's SNS topic"""
    topic_arn = create_topic_for_tag(tag_name)
    
    response = sns_client.subscribe(
        TopicArn=topic_arn,
        Protocol='email',
        Endpoint=email
    )
    
    print(f"Subscribed {email} to tag '{tag_name}'")
    return response

def notify_subscribers(tag_name, file_url, file_type):
    """Notify all subscribers when a new file with a specific tag is uploaded"""
    topic_arn = create_topic_for_tag(tag_name)
    
    message = f"""
A new {file_type} has been uploaded with tag: {tag_name}

File URL: {file_url}

You are receiving this notification because you subscribed to updates for '{tag_name}'.
To unsubscribe, click the unsubscribe link in this email.
    """
    
    response = sns_client.publish(
        TopicArn=topic_arn,
        Subject=f"New Wildlife Sighting: {tag_name.capitalize()}",
        Message=message
    )
    
    print(f"Notification sent for tag '{tag_name}': {response['MessageId']}")
    return response


def notify_tag_updated(tag_name, file_url, operation):
    """Notify subscribers when a tag is manually added or removed"""
    topic_arn = create_topic_for_tag(tag_name)
    
    action = "added to" if operation == 1 else "removed from"
    
    message = f"""
The tag '{tag_name}' has been {action} a file.

File URL: {file_url}

You are receiving this notification because you subscribed to updates for '{tag_name}'.
    """
    
    response = sns_client.publish(
        TopicArn=topic_arn,
        Subject=f"Wildlife Tag Update: {tag_name.capitalize()}",
        Message=message
    )
    
    print(f"Tag update notification sent for '{tag_name}': {response['MessageId']}")
    return response

def unsubscribe_email_from_tag(email, tag_name):
    """Unsubscribe an email address from a specific tag's SNS topic"""
    topic_arn = create_topic_for_tag(tag_name)
    
    # Find the subscription ARN for this email
    paginator = sns_client.get_paginator('list_subscriptions_by_topic')
    for page in paginator.paginate(TopicArn=topic_arn):
        for subscription in page['Subscriptions']:
            if subscription['Endpoint'] == email:
                sns_client.unsubscribe(
                    SubscriptionArn=subscription['SubscriptionArn']
                )
                print(f"Unsubscribed {email} from tag '{tag_name}'")
                return {
                    'statusCode': 200,
                    'body': json.dumps({'message': f'Unsubscribed {email} from {tag_name}'})
                }
    
    return {
        'statusCode': 404,
        'body': json.dumps({'message': 'Subscription not found'})
    }

def lambda_handler(event, context):
    """
    Lambda entry point for SNS notifications
    
    Event format for new file upload:
    {
        "action": "notify_upload",
        "tag_name": "koala",
        "file_url": "https://...",
        "file_type": "image"
    }
    
    Event format for tag update:
    {
        "action": "notify_tag_update",
        "tag_name": "koala",
        "file_url": "https://...",
        "operation": 1
    }
    
    Event format for new subscription:
    {
        "action": "subscribe",
        "email": "user@example.com",
        "tag_name": "koala"
    }
    """

    import json
    body = json.loads(event.get('body', '{}')) if isinstance(event.get('body'), str) else event
    action = body.get('action')
    
    if action == 'notify_upload':
        return notify_subscribers(
            body['tag_name'],
            body['file_url'],
            body['file_type']
    )

    elif action == 'notify_tag_update':
        return notify_tag_updated(
            body['tag_name'],
            body['file_url'],
            body['operation']
    )

    elif action == 'subscribe':
        return subscribe_email_to_tag(
            body['email'],
            body['tag_name']
    )

    elif action == 'unsubscribe':
        return unsubscribe_email_from_tag(
            body['email'],
            body['tag_name']
    )

    else:
        return {
            'statusCode': 400,
            'body': 'Invalid action'
        }