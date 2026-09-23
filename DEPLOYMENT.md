# Navidrome Serverless Deployment Guide

This guide walks you through deploying your Navidrome Serverless setup on **Vercel** with **NeonDB (PostgreSQL)** and **Cloudflare R2** for free global music streaming across all your devices (Mac, Windows, Linux, Android, iOS).

---

## Architecture Summary

* **Frontend & Subsonic API**: Hosted on **Vercel** as serverless functions.
* **Database**: **Neon Serverless PostgreSQL** managed via `sqlc` and `pgx/v5`.
* **Music Storage & Streaming**: **Cloudflare R2** with HTTP 302 presigned streaming redirects directly from Cloudflare's edge CDN ($0 egress fees, zero timeouts).
* **Sync Engine**: On-demand webhook (`/api/scan/r2`) or free automated GitHub Actions cron workflow (`.github/workflows/r2-sync.yml`).

---

## Step 1: Set Up Neon PostgreSQL (Free)

1. Sign up at [neon.tech](https://neon.tech) and create a free project (e.g. `navidrome-db`).
2. Copy your connection string:
   ```
   postgres://username:password@ep-xyz.us-east-2.aws.neon.tech/neondb?sslmode=require
   ```
3. Initialize the database schema with the consolidated schema file:
   ```bash
   psql "postgres://username:password@ep-xyz.us-east-2.aws.neon.tech/neondb?sslmode=require" -f db/postgres/schema.sql
   ```

---

## Step 2: Set Up Cloudflare R2 (Free 10 GB)

1. Log into your [Cloudflare Dashboard](https://dash.cloudflare.com/) > **R2 Object Storage**.
2. Click **Create Bucket** (e.g. `my-navidrome-music`).
3. Under **Manage R2 API Tokens**, click **Create API Token**:
   * Permissions: **Object Read & Write**
   * Copy the **Access Key ID**, **Secret Access Key**, and your **Account ID**.
4. Upload your music folders into the bucket (via Cloudflare Web UI, Cyberduck, or `rclone`).

---

## Step 3: Deploy to Vercel (Free)

1. Push this directory (`/Users/abhay/Desktop/navidrome-serverless`) to your GitHub repository:
   ```bash
   git remote add origin https://github.com/<your-username>/navidrome-serverless.git
   git branch -M main
   git push -u origin main
   ```
2. Log into [vercel.com](https://vercel.com) > **Add New** > **Project** > Import your repository.
3. In **Environment Variables**, add:

| Environment Variable | Value / Description |
| :--- | :--- |
| `NEON_DATABASE_URL` | Your Neon connection string from Step 1 |
| `ND_R2_ACCOUNTID` | Your Cloudflare Account ID |
| `ND_R2_ACCESSKEYID` | Cloudflare R2 Access Key ID |
| `ND_R2_SECRETACCESSKEY` | Cloudflare R2 Secret Access Key |
| `ND_R2_BUCKET` | Your R2 bucket name (e.g. `my-navidrome-music`) |
| `ND_R2_ENABLEPRESIGNEDSTREAM` | `true` |
| `ND_SCAN_SECRET` | A secure random token for triggering library scans |

4. Click **Deploy**. Vercel will build and provide your live URL (e.g. `https://navidrome-serverless.vercel.app`).

---

## Step 4: Indexing Your Music (R2 Sync)

Because serverless cannot run 24/7 background scanners, you can sync newly added music in three ways:

### Option A: Webhook (Anytime you add music)
Trigger an on-demand scan via curl:
```bash
curl -X POST "https://your-app.vercel.app/api/scan/r2?secret=YOUR_ND_SCAN_SECRET"
```

### Option B: Automated GitHub Actions (Runs every 6 hours)
In your GitHub repo, go to **Settings** > **Secrets and variables** > **Actions** and add the same environment variables. The workflow [`.github/workflows/r2-sync.yml`](.github/workflows/r2-sync.yml) will sync automatically every 6 hours or whenever you click "Run workflow".

### Option C: Local CLI
Run directly on your computer:
```bash
go run main.go scan-r2
```

---

## Step 5: Connect All Your Devices

Your Vercel deployment exposes the full Subsonic API:

* **Server URL**: `https://your-app.vercel.app`
* **Username & Password**: Your Navidrome admin or user account credentials.

### Recommended Client Apps:
* **All Devices (Browser)**: Open `https://your-app.vercel.app/app` directly in Safari/Chrome (PWA supported).
* **Android**: [Symfonium](https://symfonium.app/) or [DSub](https://github.com/daneren200/navidrome-dsub).
* **iOS (iPhone / iPad)**: [Substreamer](https://substreamer.app/), [Amplefor](https://amplefor.com/), or [play:Sub](https://playsub.app/).
* **macOS / Windows / Linux**: [Feishin](https://github.com/jeffvli/feishin) or [Supersonic](https://github.com/dweomer/supersonic).
