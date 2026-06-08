# Aussie EcoLens - Multi-Cloud Wildlife Observation Platform

## Overview

Aussie EcoLens is a serverless multi-cloud wildlife observation platform. Users can upload, manage, and search wildlife images and videos through a web interface. The system automatically detects species using a provided machine-learning model and stores the processed metadata for future queries.

## Multi-Cloud Architecture

The platform integrates AWS and Google Cloud Platform services.

### AWS Services

* **Amazon Cognito**: User authentication and authorisation
* **Amazon API Gateway**: Protected REST API entry point
* **AWS Lambda**: Upload URL generation and media-processing orchestration
* **Amazon S3**: Storage for images, videos, and image thumbnails
* **Amazon DynamoDB**: Metadata, detected tags, tag counts, and storage URLs
* **Amazon SNS**: Tag-based email notifications

### GCP Services

* **Cloud Run Media Inference Service**: Species detection, image processing, and video frame extraction
* **Cloud Run Go Query API**: Search, lookup, bulk tag updates, deletion, and video download support
* **Google Cloud Storage**: Storage for the provided Aussie EcoLens model artifacts

The main upload and processing flow is:

```text
Vue Frontend
→ AWS API Gateway
→ AWS Lambda
→ Amazon S3
→ S3 Event
→ AWS Lambda
→ GCP Cloud Run Media Inference
→ Amazon DynamoDB
```

Query and data-management requests follow this flow:

```text
Vue Frontend
→ AWS API Gateway
→ GCP Cloud Run Go Query API
→ Amazon DynamoDB / Amazon S3
```

## Team Members

| Name        | Student ID | Role                                     |
| ----------- | ---------: | ---------------------------------------- |
| Yichao Zhu  |   35636572 | AWS Infrastructure and Security          |
| Jackson Xia |   35817135 | Media Processing and Serverless Pipeline |
| Jiawei Wang |   35304723 | Go Query API and Data Management         |
| Boting Liu  |   35201061 | Frontend UI and Integration              |

## Main Features

* User registration, authentication, and authorisation through AWS Cognito
* Image and video upload using presigned S3 URLs
* Duplicate-upload prevention using SHA-256 checksums
* Automatic image thumbnail generation
* Video frame extraction at 1 frame per second
* ML-based wildlife species tagging
* Metadata storage in DynamoDB
* Species search and JSON tag-count queries
* Logical AND matching for multi-tag queries
* Thumbnail-to-original image URL lookup
* Temporary image search without permanent storage
* Bulk tag addition and removal
* Bulk file deletion
* Video download support
* Tag-based email notifications through SNS

## Infrastructure Configuration

### AWS

* **Region**: `us-east-1`
* **Cognito User Pool ID**: `us-east-1_zHpMn5rX5`
* **Cognito Client ID**: `4vaujh7rjc3as3q38pqlva6iet`
* **S3 Bucket**: `fit5225-wildlife-media`
* **DynamoDB Table**: `fit5225-wildlife-media-metadata`
* **API Gateway URL**: `https://aetsjr34k4.execute-api.us-east-1.amazonaws.com`

### GCP

* **Project**: `fit5225-wildlife-platform`
* **Region**: `australia-southeast1`
* **Cloud Run Media Inference**: `https://wildlens-media-infer-343888474330.australia-southeast1.run.app`
* **Cloud Run Go Query API**: `https://wildlens-query-api-343888474330.australia-southeast1.run.app`
* **Model Artifacts Bucket**: `fit5225-wildlens-model-artifacts`

## Getting Started

### Prerequisites

* AWS Academy Learner Lab access
* GCP access
* Node.js 18+
* Python 3.10+
* Go 1.22+

### Clone the Repository

```bash
git clone https://github.com/jiawei-wang-dev/wildlens-cloud.git
cd wildlens-cloud
```

### Run the Frontend

```bash
cd frontend
npm install
npm run dev
```

### Run the Query API Tests

```bash
cd backend/query-api
go test -v ./...
```

## Repository Structure

```text
wildlens-cloud/
├── README.md
├── .env.example
├── backend/
│   └── query-api/                 # Go Query API - Jiawei Wang
├── docs/                          # API documentation and diagrams
├── frontend/                      # Vue frontend - Boting Liu
└── functions/
    └── media_processing/          # Python media inference - Jackson Xia
```

## Security Notes

Do not commit AWS Academy temporary credentials, session tokens, passwords, JWT tokens, or GCP service-account private keys to the repository.

The repository must remain private and should be shared with the teaching team. Each team member should commit their own work to preserve evidence of individual contributions.
