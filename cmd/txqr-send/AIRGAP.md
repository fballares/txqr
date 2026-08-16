# Air-gapped screen → phone transfer

## Your constraints

- No network file transfer
- No USB
- Only option: show content on the Windows screen and scan with a phone

TXQR is built for exactly this: **animated QR on a monitor → camera decode → reconstruct payload**.

## How start / end works (you do NOT need frame 1)

Each QR frame looks like:

```text
blockCode/chunkLen/total|payload
```

| Field | Meaning |
|-------|---------|
| `blockCode` | Fountain-code block ID (not a strict sequence number) |
| `chunkLen` | Encoder chunk size |
| `total` | Full payload length in bytes |
| `payload` | Piece of fountain-coded data |

**Start:** There is no required “first” frame. The phone can join mid-stream. The first valid frame the decoder sees teaches it `total` and `chunkLen`.

**End:** There is no special end frame. The decoder finishes when fountain codes have enough blocks to rebuild the whole payload (`IsCompleted()` → `Data()`).

**Missed frames:** Expected. Fountain codes are designed for erasure channels (phone cameras miss frames constantly).

**Duplicates:** Ignored (header cache), so holding on one frame or reading the same frame twice is fine.

## Continuous looping (late phone start is OK)

The Windows overlay **loops forever** until you close it:

1. You copy text / trigger QR
2. Overlay appears (bottom-right by default) and keeps cycling frames
3. Whenever the phone is ready, point it at the overlay
4. Keep scanning until the mobile app reports complete
5. Close the overlay

You do **not** need to catch the beginning of the animation.

## Movable overlay for phone placement

- Overlay opens as a **small window** (about 420×540)
- On Windows: prefers **bottom-right**, tries **always-on-top**
- **Drag the window** anywhere convenient (desk corner, beside keyboard, etc.)
- Park the phone on a stand aimed at that spot for hands-free scanning

## Multi-QR concurrent streams

Windows **auto-selects** 1 vs 2 QR codes so dual only appears when it should finish faster (larger payloads / longer loops). The iPhone reader dynamically accepts whatever it sees.

- Wire format unchanged
- Override: `-streams 1` or `-streams 2`
- Sweet spot when dual is chosen: **2 streams**

## Phone scanning limits

| Factor | Practical limit |
|--------|------------------|
| Useful decode rate | Often ~2–5 unique frames/sec handheld; better on a stand |
| Animation FPS | We use ~5–6 FPS so the camera can keep up |
| QR density | Medium ECC, ~110–150 byte chunks — denser = fewer frames but harder focus |
| One QR in view | Most phone libraries return **one** code per camera frame |
| Multi-QR on screen | Possible in theory (2–4 codes) for higher bandwidth, but needs a custom reader that finds multiple codes per frame; stock apps usually don’t |
| Glare / angle / autofocus | Biggest real-world failures — matte screen / slight tilt helps |
| Session length | Comfortable for seconds–a few minutes; multi‑MB is too long for handheld |

## What stays easy vs what fights you

**Easy with this design**

- Clipboard text, notes, configs, short documents
- Starting the phone late
- Missing some frames
- Leaving the overlay looping while you position the phone

**Hard with screen-only QR**

- Multi‑megabyte Word/Excel as a full binary (hours / tens of minutes)
- Expecting a stock camera app to assemble the file (you need a TXQR-aware reader that loops decode until complete, then saves bytes)

## Mobile reader checklist

1. Continuously decode camera frames (don’t stop after one QR)
2. Feed each payload into `Decoder.Decode`
3. Ignore validation failures (other QRs in the room)
4. When `IsCompleted()`, take `Data()` / `DataBytes()` and copy or save
5. `Reset()` before the next transfer

That matches the air-gapped “show on screen, scan until done” workflow.
