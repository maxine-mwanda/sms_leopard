# SYSTEM OVERVIEW

SMSLeopard — RabbitMQ + MySQL Edition

Overview
--------
This service provides HTTP APIs to create/list campaigns, preview messages, and send campaigns to customers. It separates concerns into:

- controllers/: HTTP handlers (thin)
- models/: service + DB logic (MySQL)
- db/: DB connection and migrations
- queue/: RabbitMQ publisher & consumer helpers
- worker/: background worker consumes jobs and processes outbound messages
- cmd/server/: server entrypoint
- tests/: unit tests using sqlmock

Message flow
------------
1. Client calls `POST /campaigns` to create a campaign.
2. Client calls `POST /campaigns/send` to enqueue sending.
   - Service creates `outbound_messages` rows with status `queued`.
   - Publisher publishes a job `{campaign_id}` to RabbitMQ.
3. Worker (consumer) receives the job, renders templates, updates each outbound row with body and status `sent`, and (in production) would call an SMS provider adapter.

Data model
----------
- customers(id, phone, first_name, last_name, metadata)
- campaigns(id, name, template, status, created_at)
- outbound_messages(id, campaign_id, customer_id, to_phone, body, status, created_at)

Design notes
------------
- Template rendering: simple token substitution `{{first_name}}` for determinism and easy testing; swap to `text/template` or an AI step as needed.
- Queue: RabbitMQ for durability and delivery guarantees; publisher uses exchange, consumer uses durable queue.
- Tests: Use sqlmock to test DB interactions; queue publish/consume logic accepted but consumer behavior tested via unit tests and mocks.

