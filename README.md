<a href="https://github.com/QUBITABHAY/navidrome-serverless"><img src="resources/logo-192x192.png" alt="Navidrome Serverless logo" title="Navidrome Serverless" align="right" height="70px" /></a>

# Navidrome Serverless 🚀

**Navidrome Serverless** is a cloud-native, serverless distribution of [Navidrome](https://www.navidrome.org). It is re-architected to run entirely on **100% free-tier services**—deployed as serverless functions on **Vercel**, backed by any **Serverless or Managed PostgreSQL** database, and streaming audio directly from **Cloudflare R2** / AWS S3.

> [!NOTE]
> **No VPS or 24/7 server required!** Enjoy your personal Spotify-like streaming service with **zero idle compute cost**, scale-to-zero database connection pooling, and **$0 egress fees** on Cloudflare R2.

---

## 🌟 Key Differences from Upstream Navidrome

Standard Navidrome is designed for persistent virtual machines or home servers with a local filesystem and SQLite database. **Navidrome Serverless** adapts the core architecture for stateless, ephemeral environments:

| Feature                  | Standard Navidrome              | Navidrome Serverless                                                                                        |
| :----------------------- | :------------------------------ | :---------------------------------------------------------------------------------------------------------- |
| **Hosting**              | VPS, Docker, Raspberry Pi       | **Vercel Serverless Functions** (or AWS Lambda)                                                             |
| **Database**             | Local SQLite (`navidrome.db`)   | **External / Serverless PostgreSQL** via type-safe [`sqlc`](https://sqlc.dev) & `pgx/v5` connection pooling |
| **Audio Storage**        | Local hard drive / NFS / SMB    | **Cloudflare R2** or AWS S3 Object Storage                                                                  |
| **Streaming Delivery**   | Server proxies all audio chunks | **Direct HTTP 302 Presigned Streaming** from Cloudflare Edge CDN ($0 egress, zero function timeouts)        |
| **Library Scanner**      | Long-running background daemon  | **On-demand Webhook**, automated **GitHub Actions Cron**, or local CLI (`scan-r2`)                          |
| **Client Compatibility** | Subsonic API + Web UI           | **100% Identical** (Symfonium, Substreamer, Feishin, DSub, Web UI, etc.)                                    |

---

## 🏗️ Architecture

```mermaid
flowchart TD
    Client["Mobile & Desktop Clients\n(Symfonium, Substreamer, Feishin, Web)"]
    Vercel["Vercel Serverless Function\n(Go Handler / Subsonic API)"]
    Postgres[("PostgreSQL Database\n(Metadata via sqlc)")]
    R2[("Cloudflare R2 / S3\n(Audio Files)")]
    Scanner["Library Sync Engine\n(GitHub Actions / Webhook / CLI)"]

    Client -->|"API requests (browse, playlists, search)"| Vercel
    Vercel <-->|"Scale-to-zero pgx pool"| Postgres
    Client -->|"GET /rest/stream"| Vercel
    Vercel -->|"HTTP 302 Redirect (Presigned URL)"| Client
    Client -->|"Direct audio playback ($0 egress)"| R2
    Scanner -->|"HTTP Range scan metadata (ID3/Vorbis)"| R2
    Scanner -->|"Update song/album catalog"| Postgres
```

---

## 📱 Client Compatibility

Navidrome Serverless implements the standard Subsonic API (v1.16.1) and is fully compatible with any Subsonic client across all platforms:

- **Web Browser**: Built-in responsive React / Material UI web player (`https://your-domain.vercel.app/app`)
- **iOS (iPhone, iPad, CarPlay)**: [Substreamer](https://substreamer.app/), [Amplefor](https://amplefor.com/), [play:Sub](https://playsub.app/)
- **Android (Auto, WearOS)**: [Symfonium](https://symfonium.app/), [DSub](https://github.com/daneren200/navidrome-dsub)
- **macOS / Windows / Linux**: [Feishin](https://github.com/jeffvli/feishin), [Supersonic](https://github.com/dweomer/supersonic)

---

## 🚀 Quick Start & Deployment

For detailed step-by-step instructions, see the complete [**Deployment Guide (DEPLOYMENT.md)**](DEPLOYMENT.md).

### 1. Set up PostgreSQL

1. Create a PostgreSQL database using any provider of your choice (e.g. Neon, Supabase, Tembo, Aiven, or self-hosted).
2. Apply the consolidated PostgreSQL schema:
   ```bash
   psql "YOUR_POSTGRES_DATABASE_URL" -f db/postgres/schema.sql
   ```

### 2. Set up Cloudflare R2 (or AWS S3)

1. Create an R2 bucket in the Cloudflare Dashboard (e.g., `my-music`).
2. Generate an R2 API token with **Object Read & Write** permissions.
3. Upload your music folders into the bucket.

### 3. Deploy to Vercel

1. Import your GitHub repository to [Vercel](https://vercel.com).
2. Configure the following environment variables:
   - `DATABASE_URL`: `postgres://user:password@hostname:5432/dbname?sslmode=require`
   - `ND_R2_ACCOUNTID`: Your Cloudflare Account ID
   - `ND_R2_ACCESSKEYID`: Your Cloudflare R2 Access Key ID
   - `ND_R2_SECRETACCESSKEY`: Your Cloudflare R2 Secret Access Key
   - `ND_R2_BUCKET`: Your bucket name (e.g. `my-music`)
   - `ND_R2_ENABLEPRESIGNEDSTREAM`: `true`
   - `ND_SCAN_SECRET`: Secret token for webhook scans
3. Click **Deploy**.

### 4. Index Your Music Library

You can trigger metadata extraction in 3 convenient ways:

- **Automated Cron**: Included GitHub Actions workflow ([`.github/workflows/r2-sync.yml`](.github/workflows/r2-sync.yml)) scans every 6 hours automatically.
- **On-Demand Webhook**:
  ```bash
  curl -X POST "https://your-app.vercel.app/api/scan/r2?secret=YOUR_ND_SCAN_SECRET"
  ```
- **Local CLI**:
  ```bash
  go run main.go scan-r2
  ```

---

## 🛠️ Local Development & Testing

```bash
# Clone the repository
git clone https://github.com/QUBITABHAY/navidrome-serverless.git
cd navidrome-serverless

# Run code linters
make lint

# Run unit tests
make test

# Generate sqlc models (if modifying PostgreSQL queries)
sqlc generate
```

---

## 📜 Credits & Acknowledgements

- Original [Navidrome](https://github.com/navidrome/navidrome) project created and maintained by [Deluan Quintao](https://github.com/deluan) and contributors.
- Navidrome Serverless maintains Subsonic API specification compatibility while decoupling compute, database, and storage for modern cloud architectures.

---

## 📄 License

Navidrome Serverless is licensed under the [GNU General Public License v3.0 (GPLv3)](LICENSE).
