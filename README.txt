README.txt

Project: Price Tracker

This repository contains two main components:
 1. Backend service (Go + Python)
 2. Frontend application (Flutter for Windows)

── BACKEND (Go + Python) ────────────────────────────────────────────────────────

Folder structure:
  Price_tracker/
    ├── main.go
    ├── go.mod
    ├── go.sum
    ├── fetch_price.py
    ├── requirements.txt
    └── cache.db                ← created at runtime

Prerequisites:
 • Go 1.24+ (amd64)
 • Python 3.8+
 • Tesseract OCR installed and in PATH
 • (Windows) Git in PATH

Setup & Run:

1. Open PowerShell and navigate to the backend directory:
     cd C:\path\to\price_tracker

2. Build Go executable:
     go mod tidy
     go build -o Price_tracker.exe main.go

3. Prepare Python virtual environment:
     python -m venv venv
     .\venv\Scripts\Activate

4. Install Python dependencies:
     pip install -r requirements.txt

5. Run the backend service:
     .\Price_tracker.exe

   The service listens on http://localhost:8080
   Endpoints:
     /api/price?asin=<ASIN>    → returns JSON { asin, highest, lowest, current, average, fetched_at }
     /api/image?asin=<ASIN>    → returns full chart PNG (cached 30 min)

── FRONTEND (Flutter) ─────────────────────────────────────────────────────────

Folder structure:
  Tracker_app/
    ├── pubspec.yaml
    └── lib/
        └── main.dart

Prerequisites:
 • Flutter SDK (stable channel)
 • Visual Studio Build Tools (with “Desktop development with C++”)
 • PowerShell or CMD
 • (Optional) Android Studio if you want mobile

Setup & Run:

1. Open PowerShell and navigate to the Flutter project:
     cd C:\path\to\rastreador_precos_app

2. Fetch Dart dependencies:
     flutter pub get

3. Enable Windows desktop support (one-time):
     flutter config --enable-windows-desktop

4. Verify devices:
     flutter devices
     → should list “Windows (desktop)”

5. Run the app on Windows:
     flutter run -d windows

   The app will open a native window.  
   Enter an Amazon product URL or ASIN, click “Check Price”,  
   and you will see the price data and chart image from your backend.

── NOTES ───────────────────────────────────────────────────────────────────────

 • Make sure the backend service is running before starting the Flutter app.
 • If you deploy the Go service to a remote server, update kApiBase in lib/main.dart.
 • To rebuild after code changes:
     – Backend: re-run `go build` and restart the .exe
     – Frontend: hot-reload with “r” in the flutter console or run again

Happy coding!
