# Motivations & Decisions

This project started as an experiment on how to use AI and Spec-Driven Development to develop and deploy a (relatively) simple application.

## Why cocktails?

I enjoy making cocktails and I keep a small notebook with recipies that I have enjpoyed as well as variations that I have made over time. Sometimes I forget to write the recipe in my notebook and I need to start searching. Other times I remember a drink ingredient but it is not listed in the index of any of my books, or want to use an ingredient and, again, it is not listed. 

Whith that in mind here are the goals of this application:

1. Collect recipies and variations of cocktails
2. Make the recipies searchable by any ingredient
3. Add extra tags and descriptions that can be searched
4. Allow friends to login and add recipies or simply browse them

## Technology choices

Another goal was to use new technologies (AI and Spec-kit) and see how I can create programs in languages I am not familiar with, and to also incorporate techniques, frameworks and languages that I already know and I am familiar with (Terraform, SAM, AWS Lambda, etc.) 

It was important that I would do as little programming as possible, or hopefully none. _The whole point of the experiment_ was to create a 100% AI driven project.

Here is a summary of the choices I have made:

#### AI
- Claude Code for programming
- Spec-kit for specification
- Terrashark skill to create Terraform code

#### Languages:
- Go for the backend and lambda
- JavaScript/Node for the frontend
- Terraform and/or SAM for infrastructure deployment

#### Frameworks
- Vite and tailwind for frontend development
- [Terraform Serverless](https://serverless.tf/)

#### Databases
- DynamoDB everywhere — the AWS-managed service in production, and a DynamoDB Local container for local development and tests (originally SQLite locally, replaced in feature 029 for parity with production)

#### Cloud and services
- AWS
- DynamoDB
- Lambda
- API Gateway
- S3
- CloudFront

