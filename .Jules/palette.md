## 2024-05-22 - Copy Interaction Feedback
**Learning:** Users experience anxiety when copying sensitive data (like API keys) that won't be shown again. Visual confirmation ("Copied!" + icon) reduces this cognitive load significantly compared to silent clipboard writes.
**Action:** Always wrap `navigator.clipboard.writeText` with a visual state change (text update, icon change, or toast) that persists for ~2 seconds.
