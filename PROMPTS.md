# PROMPTS

Raw, unedited prompts used with the AI coding assistant (Cursor, Claude), in
order. Multiple-choice answers to clarifying questions are noted as well.

---

## Prompt 1 — task framing / discussion

> hey bot, I have task with input is a large csv file ~1GB. I already downloaded it and this is the task description below, please summarize what I need to be done. Do not code right now, please discuss abotu solution, trade off and architecture:
> CSV Schema
> Column	Type	Description
> campaign_id	string	Campaign ID
> date	string	Date in YYYY-MM-DD format
> impressions	integer	Number of impressions
> clicks	integer	Number of clicks
> spend	float	Advertising cost (USD)
> conversions	integer	Number of conversions
> Example:
> campaign_id	date	impressions	clicks	spend	conversions
> CMP001	2025-01-01	12000	300	45.50	12
> CMP002	2025-01-01	8000	120	28.00	4
> CMP001	2025-01-02	14000	340	48.20	15
> CMP003	2025-01-01	5000	60	15.00	3
> CMP002	2025-01-02	8500	150	31.00	5
> 🎯 Task Requirements
> You must build a console application (CLI) in any programming language (Python, NodeJS, Go, Java, Rust, etc.) that processes the CSV file and produces aggregated analytics.
>
> 1. Aggregate data by campaign_id
> For each campaign_id, compute:
>
> total_impressions
> total_clicks
> total_spend
> total_conversions
> CTR = total_clicks / total_impressions
> CPA = total_spend / total_conversions
> If conversions = 0, ignore or return null for CPA
> 2. Generate two result lists
> A. Top 10 campaigns with the highest CTR
> Output as CSV format.
> ...
> B. Top 10 campaigns with the lowest CPA
> Output as CSV format. Exclude campaigns with zero conversions.
> ...
> 3. Technical Requirements
> The file is large (~1GB).
> Your solution must handle large datasets efficiently with good performance and memory optimization.
> The program should be runnable via CLI, for example: python aggregator.py --input ad_data.csv --output results/
> ...
> ✅ Code Quality Expectations
> ... Correct results, Clean readable code, Error handling, Performance awareness, Tests, Documented decisions ...

(Full task description as provided in the assignment.)

### Clarifying answers given to the assistant

- Language/engine: **Go**
- Output format: **Comma-separated CSV, CTR 4 decimals, CPA 2 decimals**
- Optional extras: **Everything (Dockerfile + synthetic data generator + benchmark logs + PROMPTS.md)**
- CSV parsing strategy: **Use `encoding/csv` (robust quote/escape handling)**

---

## Prompt 2 — implementation

> Go CSV Ad Aggregator
>
> Implement the plan as specified, it is attached for your reference. Do NOT edit the plan file itself.
>
> To-do's from the plan have already been created. Do not create them again. Mark them as in_progress as you work, starting with the first one. Don't stop until you have completed all the to-dos.
