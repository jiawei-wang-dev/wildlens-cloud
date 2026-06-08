# Aussie EcoLens - Multi-Cloud Wildlife Observation Platform

## Overview
A serverless multi-cloud platform that allows users to upload, tag, and search wildlife media files. The system uses machine learning to automatically identify species in images and videos.

## Architecture
- **AWS**: Authentication (Cognito), Storage (S3), Notifications (SNS), API Gateway, Lambda, DynamoDB
- **GCP**: Media Processing (Cloud Run), ML Inference

## Team Members
| Name | Student ID | Role |
|------|-----------|------|
| [Member A] | [ID] | AWS Infrastructure & Security |
| [Member B] | [ID] | Media Processing & Serverless |
| [Member C] | [ID] | Query API & Data Management |
| [Member D] | [ID] | Frontend UI |

## Features
- User authentication and authorisation via AWS Cognito
- Media file upload with automatic species tagging
- Image thumbnail generation
- Video frame extraction for species detection
- Tag-based search and queries
- Email notifications via AWS SNS
- Manual tag management
- Duplicate file detection via SHA256 checksum
- Multi-cloud pipeline: AWS Lambda → GCP Cloud Run → AWS DynamoDB

## AWS Infrastructure
- **Cognito User Pool ID**: us-east-1_zHpMn5rX5
- **Cognito Client ID**: 4vaujh7rjc3as3q38pqlva6iet
- **S3 Bucket**: fit5225-wildlife-media
- **DynamoDB Table**: fit5225-wildlife-media-metadata
- **API Gateway URL**: https://aetsjr34k4.execute-api.us-east-1.amazonaws.com
- **Region**: us-east-1

## GCP Infrastructure
- **Project**: fit5225-wildlife-platform
- **Region**: australia-southeast1
- **Cloud Run ML**: https://wildlens-media-infer-343888474330.australia-southeast1.run.app

## Getting Started
### Prerequisites
- AWS Academy account
- GCP account
- Node.js 18+
- Python 3.10+

### Installation
```bash
git clone https://github.com/jiawei-wang-dev/wildlens-cloud.git
cd wildlens-cloud
```

## Repository Structure
```
wildlens-cloud/
├── README.md
├── aws-config.md
├── frontend/          (BotingLiu 35201061)
├── media-processing/  (JacksonXia 35817135)
├── query-api/         (JiaWei Wang 35304723)
└── auth/              (Yichao Zhu 35636572)
