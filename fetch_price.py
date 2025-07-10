import argparse
import json
import re
from datetime import datetime
from io import BytesIO

import pytesseract
import requests
from PIL import Image, ImageEnhance, ImageFilter

pytesseract.pytesseract.tesseract_cmd = r"C:\Program Files\Tesseract-OCR\tesseract.exe"

def baixar_e_extrair_valores(asin: str):
    """
    Downloads CamelCamelCamel charts (amazon.png and new.png),
    applies preprocessing and OCR, and returns a list of floats.
    """
    urls = [
        f"https://charts.camelcamelcamel.com/us/{asin}/amazon.png?force=1&zero=0&w=953&h=494.5&desired=false&legend=1&ilt=1&tp=6m&fo=0&lang=en",
        f"https://charts.camelcamelcamel.com/us/{asin}/new.png?force=1&zero=0&w=953&h=494.5&desired=false&legend=1&ilt=1&tp=6m&fo=0&lang=en",
    ]

    for url in urls:
        resp = requests.get(url, headers={"User-Agent": "Mozilla/5.0"}, timeout=10)
        if resp.status_code != 200 or "image" not in resp.headers.get("Content-Type", ""):
            continue

        img = Image.open(BytesIO(resp.content))
        w, h = img.size

        left = max(0, w - 300)
        crop = img.crop((left, 0, w, h))

        crop = crop.convert("L")
        crop = crop.resize((crop.width * 3, crop.height * 3), Image.LANCZOS)
        crop = crop.filter(ImageFilter.MedianFilter(3))
        crop = crop.point(lambda x: 0 if x < 200 else 255)

        custom_config = r'--psm 6 -c tessedit_char_whitelist=0123456789.,$'
        texto = pytesseract.image_to_string(crop, config=custom_config)

        encontrados = re.findall(r"\$?[\d,]+\.\d{2}", texto)
        valores = sorted(
            float(v.replace('$', '').replace(',', ''))
            for v in encontrados
        )
        if valores:
            return valores

    raise RuntimeError(f"Unable to extract values ​​for ASIN {asin}")

def compute_price_data(asin: str):
    """
    Runs OCR in memory and returns a dict with:
    asin, highest, lowest, current, average, fetched_at
    """
    valores = baixar_e_extrair_valores(asin)

    if len(valores) == 1:
        high = low = current = valores[0]
        avg = current
    elif len(valores) == 2:
        high, low = max(valores), min(valores)
        current = (high + low) / 2
        avg = current
    else:
        high, low = max(valores), min(valores)
        ordered = sorted(valores)
        current = ordered[1]
        avg = (high + low) / 2

    return {
        "asin":      asin,
        "highest":   high,
        "lowest":    low,
        "current":   current,
        "average":   avg,
        "fetched_at": datetime.utcnow().isoformat() + "Z"
    }

def main():
    parser = argparse.ArgumentParser(
        description="Extracts prices from CamelCamelCamel chart for an ASIN and generates JSON"
    )
    parser.add_argument("--asin", required=True, help="Código ASIN do produto")
    args = parser.parse_args()

    try:
        data = compute_price_data(args.asin)
        print(json.dumps(data))
    except Exception as e:
        import sys
        print(f"ERRO: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
