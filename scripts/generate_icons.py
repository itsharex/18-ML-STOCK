#!/usr/bin/env python3
"""
Generate app icons for Windows (.ico) and macOS (.icns / AppIcon.iconset)
from a source image.
"""
import os
import sys
import struct
from io import BytesIO
from PIL import Image, ImageDraw, ImageFilter

# Paths
PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC_IMG = os.path.join(PROJECT_ROOT, "assets", "icons", "source", "red-w.png")
OUT_DIR = os.path.join(PROJECT_ROOT, "assets", "icons", "generated")
WIN_ICON = os.path.join(PROJECT_ROOT, "build", "windows", "icon.ico")
DARWIN_DIR = os.path.join(PROJECT_ROOT, "build", "darwin", "AppIcon.iconset")
APPICON = os.path.join(PROJECT_ROOT, "build", "appicon.png")

# Ensure output dirs exist
os.makedirs(OUT_DIR, exist_ok=True)
os.makedirs(DARWIN_DIR, exist_ok=True)
os.makedirs(os.path.dirname(WIN_ICON), exist_ok=True)

# Icon sizes
WINDOWS_SIZES = [16, 20, 24, 32, 40, 48, 64, 96, 128, 256, 512]
MACOS_SIZES = [16, 32, 64, 128, 256, 512, 1024]


def rounded_rect_mask(size, radius_ratio=0.22):
    """Create a rounded rectangle mask for macOS-style icons."""
    w, h = size
    radius = int(min(w, h) * radius_ratio)
    mask = Image.new("L", size, 0)
    draw = ImageDraw.Draw(mask)
    draw.rounded_rectangle([0, 0, w - 1, h - 1], radius=radius, fill=255)
    return mask


def make_radial_gradient(size, center_color, edge_color):
    """Create a radial gradient image (center bright -> edge dark)."""
    w, h = size
    cx, cy = w / 2, h / 2
    max_dist = ((cx) ** 2 + (cy) ** 2) ** 0.5
    img = Image.new("RGBA", size)
    pixels = img.load()
    for y in range(h):
        for x in range(w):
            dist = ((x - cx)**2 + (y - cy)**2) ** 0.5
            t = min(1.0, dist / max_dist)
            r = int(center_color[0] * (1 - t) + edge_color[0] * t)
            g = int(center_color[1] * (1 - t) + edge_color[1] * t)
            b = int(center_color[2] * (1 - t) + edge_color[2] * t)
            a = int(center_color[3] * (1 - t) + edge_color[3] * t)
            pixels[x, y] = (r, g, b, a)
    return img


def create_bg(size, for_macos=False):
    """Create a deep blue radial gradient background with rounded corners."""
    w, h = size
    rr = 0.22 if for_macos else 0.18
    
    center_color = (50, 75, 145, 255)
    edge_color = (6, 10, 25, 255)
    bg = make_radial_gradient((w, h), center_color, edge_color)
    
    mask = rounded_rect_mask((w, h), radius_ratio=rr)
    bg.putalpha(mask)
    
    return bg


def prepare_source():
    """Load source, remove white background, upscale for smooth edges."""
    src = Image.open(SRC_IMG).convert("RGBA")
    
    # Upscale for smooth edges
    upscale = 8
    src = src.resize((src.width * upscale, src.height * upscale), Image.LANCZOS)
    
    # Remove white background
    pixels = src.load()
    for y in range(src.height):
        for x in range(src.width):
            r, g, b, a = pixels[x, y]
            if r >= 245 and g >= 245 and b >= 245:
                pixels[x, y] = (r, g, b, 0)
    
    return src


def generate_icon_design(size, for_macos=False, src=None):
    """Generate a single icon at the given size."""
    if src is None:
        src = prepare_source()
    
    w = h = size
    padding = int(size * 0.15)  # 15% padding for balanced W size
    target_w = w - padding * 2
    target_h = h - padding * 2
    
    # Preserve aspect ratio
    src_ratio = src.width / src.height
    if target_w / target_h > src_ratio:
        new_h = target_h
        new_w = int(new_h * src_ratio)
    else:
        new_w = target_w
        new_h = int(new_w / src_ratio)
    
    # Resize to target
    w_img = src.resize((new_w, new_h), Image.LANCZOS)
    
    # Small icons: extra sharpening to keep strokes solid at tiny sizes
    if size <= 64:
        w_img = w_img.filter(ImageFilter.UnsharpMask(radius=2, percent=250, threshold=3))
    
    # Background
    bg = create_bg((w, h), for_macos=for_macos)
    
    # Vertical nudge for optical center
    nudge = max(1, size // 150) if size >= 128 else 0
    
    # For macOS, add subtle outer shadow to the whole icon
    if for_macos and size >= 64:
        shadow_pad = max(6, size // 70)
        blur_r = max(6, size // 50)
        full_shadow = Image.new("RGBA", (w + shadow_pad * 2, h + shadow_pad * 2), (0, 0, 0, 0))
        full_shadow.paste((0, 0, 0, 35), (shadow_pad, shadow_pad + max(2, size // 150)), rounded_rect_mask((w, h), radius_ratio=0.22))
        full_shadow = full_shadow.filter(ImageFilter.GaussianBlur(blur_r))
        
        canvas = Image.new("RGBA", full_shadow.size, (0, 0, 0, 0))
        canvas.paste(full_shadow, (0, 0), full_shadow)
        canvas.paste(bg, (shadow_pad, shadow_pad), bg)
        
        # Center W
        wx, wy = w_img.size
        cx = (canvas.width - wx) // 2
        cy = (canvas.height - wy) // 2 - nudge
        canvas.paste(w_img, (cx, cy), w_img)
        return canvas
    else:
        final = Image.new("RGBA", (w, h), (0, 0, 0, 0))
        final.paste(bg, (0, 0), bg)
        wx, wy = w_img.size
        cx = (w - wx) // 2
        cy = (h - wy) // 2 - nudge
        final.paste(w_img, (cx, cy), w_img)
        return final


def generate_macos_iconset():
    """Generate macOS AppIcon.iconset folder."""
    print("Generating macOS AppIcon.iconset...")
    src = prepare_source()
    for size in MACOS_SIZES:
        img = generate_icon_design(size, for_macos=True, src=src)
        path = os.path.join(DARWIN_DIR, f"icon_{size}x{size}.png")
        img.save(path, "PNG")
        if size < 1024:
            path2x = os.path.join(DARWIN_DIR, f"icon_{size}x{size}@2x.png")
            img2x = generate_icon_design(size * 2, for_macos=True, src=src)
            img2x.save(path2x, "PNG")
    print(f"  -> {DARWIN_DIR}")


def generate_windows_ico():
    """Generate Windows .ico file with multiple resolutions."""
    print("Generating Windows icon.ico...")
    src = prepare_source()
    images = []
    for size in WINDOWS_SIZES:
        img = generate_icon_design(size, for_macos=False, src=src)
        images.append(img)
    
    # ICONDIR
    icondir = struct.pack('<HHH', 0, 1, len(images))
    
    # Calculate offsets
    header_size = 6 + 16 * len(images)
    entries = []
    png_datas = []
    offset = header_size
    
    for img in images:
        buf = BytesIO()
        img.save(buf, format='PNG')
        data = buf.getvalue()
        png_datas.append(data)
        w, h = img.size
        entries.append(struct.pack(
            '<BBBBHHII',
            w if w < 256 else 0,
            h if h < 256 else 0,
            0, 0, 1, 32,
            len(data), offset
        ))
        offset += len(data)
    
    with open(WIN_ICON, 'wb') as f:
        f.write(icondir)
        for entry in entries:
            f.write(entry)
        for data in png_datas:
            f.write(data)
    
    print(f"  -> {WIN_ICON}")


def generate_appicon():
    """Generate the main appicon.png for Wails."""
    print("Generating build/appicon.png...")
    src = prepare_source()
    img = generate_icon_design(1024, for_macos=True, src=src)
    img.save(APPICON, "PNG")
    print(f"  -> {APPICON}")


def main():
    if not os.path.exists(SRC_IMG):
        print(f"Source image not found: {SRC_IMG}")
        sys.exit(1)
    
    print(f"Source: {SRC_IMG}")
    generate_macos_iconset()
    generate_windows_ico()
    generate_appicon()
    print("\nDone! All icons generated.")


if __name__ == "__main__":
    main()
