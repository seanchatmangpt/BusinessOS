# Icon Styles Quick Reference

## Quick Stats

- **Total Styles**: 41
- **Modern**: 14 styles
- **Classic**: 8 styles
- **Creative**: 19 styles
- **Animated**: 4 styles (aurora, holographic, rainbow, glitch)
- **Dark Mode Variants**: 5 styles (neumorphism, fluent, chrome, sketch, comic)

---

## Style Categories At-a-Glance

### Modern (14)
```
default, minimal, rounded, square, macos,
glassmorphism, frosted, flat, paper, depth,
neumorphism, material, fluent, aero
```

### Classic (8)
```
macos-classic, retro, win95, pixel,
ios, android, windows11, amiga
```

### Creative (19)
```
outlined, neon, gradient, glow, terminal, brutalist,
aurora, crystal, holographic, vaporwave, cyberpunk,
synthwave, matrix, glitch, chrome, rainbow,
sketch, comic, watercolor
```

---

## File Locations

| Component | File | Lines |
|-----------|------|-------|
| **Type Definition** | `src/lib/stores/desktopStore.ts` | 529-589 |
| **Desktop Icons** | `src/lib/components/desktop/DesktopIcon.svelte` | 1062-1296 |
| **Dock Icons** | `src/lib/components/desktop/Dock.svelte` | 2660-2892 |
| **Settings Preview** | `src/lib/components/desktop/DesktopSettingsContent.svelte` | 2597-2924 |
| **Category Config** | `src/lib/components/desktop/DesktopSettingsContent.svelte` | 64-68 |

---

## CSS Selector Patterns

### Desktop Icons
```css
.desktop-icon.style-{styleId} .icon-image { }
.desktop-icon.style-{styleId}:hover .icon-image { }
.desktop-icon.style-{styleId} .icon-svg { }
```

### Dock Icons
```css
.dock-item.style-{styleId} .dock-icon { }
.dock-item.style-{styleId}:hover .dock-icon { }
```

### Preview Icons
```css
.style-icon.preview-{styleId} { }
:global(.dark) .style-icon.preview-{styleId} { }
```

**⚠️ CRITICAL**: Preview uses `.style-icon.preview-{styleId}` (both classes on same element), NOT `.preview-{styleId} .style-icon` (parent-child).

---

## Quick Add New Style Checklist

- [ ] 1. Add to `IconStyle` type union in `desktopStore.ts`
- [ ] 2. Add style definition object to `iconStyles` array
- [ ] 3. Add CSS to `DesktopIcon.svelte` (base + hover)
- [ ] 4. Add CSS to `Dock.svelte` (base + hover)
- [ ] 5. Add CSS to `DesktopSettingsContent.svelte` preview
- [ ] 6. Add to category in `styleCategories` object
- [ ] 7. Test: Desktop icons display correctly
- [ ] 8. Test: Dock icons display correctly
- [ ] 9. Test: Preview shows unique appearance
- [ ] 10. Test: Dark mode (if applicable)

---

## Common CSS Properties by Style Type

### Standard Gradient
```css
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
border-radius: 10px !important;
```

### Glass/Blur Effects
```css
background: rgba(255, 255, 255, 0.2) !important;
backdrop-filter: blur(10px) !important;
border: 1px solid rgba(255, 255, 255, 0.4) !important;
```

### Neon/Glow Effects
```css
background: #1a1a2e !important;
border: 1.5px solid #667eea !important;
box-shadow: 0 0 10px #667eea, 0 0 20px #667eea !important;
```

### Retro/Pixelated
```css
background: #C0C0C0 !important;
border: 2px solid !important;
border-color: #DFDFDF #808080 #808080 #DFDFDF !important;
image-rendering: pixelated !important;
```

---

## Animation Examples

### Gradient Shift
```css
.style-icon.preview-aurora {
  background: linear-gradient(135deg, #667eea, #764ba2, #f093fb) !important;
  background-size: 200% 200% !important;
  animation: aurora-shimmer 3s ease-in-out infinite !important;
}

@keyframes aurora-shimmer {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}
```

### Hue Rotation
```css
.style-icon.preview-holographic {
  background: linear-gradient(135deg, #ff0080, #ff8c00, #40e0d0, #ff0080) !important;
  animation: holographic 2s ease infinite !important;
}

@keyframes holographic {
  0% { background-position: 0% 50%; filter: hue-rotate(0deg); }
  100% { background-position: 400% 50%; filter: hue-rotate(360deg); }
}
```

### Transform Glitch
```css
.style-icon.preview-glitch {
  animation: glitch 1s infinite !important;
}

@keyframes glitch {
  0%, 100% { transform: translate(0); }
  20% { transform: translate(-2px, 2px); }
  40% { transform: translate(-2px, -2px); }
  60% { transform: translate(2px, 2px); }
  80% { transform: translate(2px, -2px); }
}
```

---

## Dark Mode Pattern

```css
/* Light mode */
.style-icon.preview-neumorphism {
  background: #e0e0e0 !important;
  box-shadow: 8px 8px 16px #bebebe, -8px -8px 16px #ffffff !important;
}

/* Dark mode */
:global(.dark) .style-icon.preview-neumorphism {
  background: #2a2a2a !important;
  box-shadow: 8px 8px 16px #1a1a1a, -8px -8px 16px #3a3a3a !important;
}
```

---

## Debugging Checklist

### Preview icons not showing correct style
- [ ] Check CSS selector: `.style-icon.preview-{styleId}` NOT `.preview-{styleId} .style-icon`
- [ ] Ensure `!important` on background property
- [ ] Verify style ID matches exactly (no typos)
- [ ] Check browser dev tools for selector specificity

### Dock icons don't match desktop icons
- [ ] Verify CSS exists in `Dock.svelte`
- [ ] Compare visual properties (background, border, border-radius)
- [ ] Check hover states match
- [ ] Test with multiple icons in dock

### Animation not working
- [ ] Verify `@keyframes` definition exists
- [ ] Check `animation` property syntax
- [ ] Ensure `background-size` set for gradient animations
- [ ] Test in different browsers

### Style not available in settings
- [ ] Added to `IconStyle` type union
- [ ] Added to `iconStyles` array
- [ ] Added to `styleCategories` object
- [ ] Check spelling consistency across all files

---

## Visual Testing Matrix

Test each new style:

| Component | State | Check |
|-----------|-------|-------|
| Desktop Icon | Normal | ✓ Visual matches design |
| Desktop Icon | Hover | ✓ Hover effect works |
| Desktop Icon | Selected | ✓ Selection indicator visible |
| Dock Icon | Normal | ✓ Matches desktop icon |
| Dock Icon | Hover | ✓ Hover effect works |
| Dock Icon | Active | ✓ Active indicator visible |
| Preview | Normal | ✓ Distinct from other styles |
| Preview | Light Mode | ✓ Visible and clear |
| Preview | Dark Mode | ✓ Dark variant (if exists) |
| Animation | Playing | ✓ Smooth and performant |

---

## Common Mistakes to Avoid

❌ **Using parent-child selector for preview**
```css
.preview-terminal .style-icon { }  /* WRONG */
```

✅ **Use same-element selector**
```css
.style-icon.preview-terminal { }  /* CORRECT */
```

---

❌ **Forgetting !important on background**
```css
.style-icon.preview-custom {
  background: #fff;  /* Gets overridden */
}
```

✅ **Add !important to override base**
```css
.style-icon.preview-custom {
  background: #fff !important;  /* Works */
}
```

---

❌ **Inconsistent styling across components**
```css
/* Desktop */
.desktop-icon.style-custom .icon-image {
  background: red;
}

/* Dock */
.dock-item.style-custom .dock-icon {
  background: blue;  /* Different! */
}
```

✅ **Match visual properties**
```css
/* Both use same colors */
background: red !important;
```

---

❌ **Missing hover states**
```css
.dock-item.style-custom .dock-icon {
  background: red;
}
/* No hover defined */
```

✅ **Define hover for better UX**
```css
.dock-item.style-custom:hover .dock-icon {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(255, 0, 0, 0.3);
}
```

---

## Performance Tips

### Efficient
✅ CSS-only animations (transform, opacity)
✅ Hardware-accelerated properties
✅ Scoped animations per-style

### Potentially Heavy
⚠️ `backdrop-filter: blur()` - Can be expensive on mobile
⚠️ Multiple box-shadows - Use sparingly
⚠️ Complex gradients with many stops

### Optimize
- Limit backdrop-filter to necessary styles only
- Use `will-change` for animated elements
- Test on target devices

---

## Icon Style State Flow

```
User opens Settings
      ↓
Select category (Modern/Classic/Creative)
      ↓
Click style preview
      ↓
handleIconStyleChange(styleId)
      ↓
desktopSettings.iconStyle = styleId
      ↓
Store updates (reactive)
      ↓
All components re-render
      ↓
✓ Desktop icons update
✓ Dock icons update
✓ Preview highlights selected
```

---

## Support & Documentation

- **Full Documentation**: [ICON_STYLES.md](./ICON_STYLES.md)
- **Changelog**: [CHANGELOG_ICON_STYLES.md](./CHANGELOG_ICON_STYLES.md)
- **Component Docs**: See individual component files

---

## Quick Code Snippets

### Get current icon style
```typescript
import { desktopSettings } from '$lib/stores/desktopStore';

const currentStyle = $desktopSettings.iconStyle;
```

### Change icon style programmatically
```typescript
import { desktopSettings } from '$lib/stores/desktopStore';

desktopSettings.update(s => ({
  ...s,
  iconStyle: 'cyberpunk'
}));
```

### Check if style is animated
```typescript
const animatedStyles = ['aurora', 'holographic', 'rainbow', 'glitch'];
const isAnimated = animatedStyles.includes(currentStyle);
```

### Filter styles by category
```typescript
import { iconStyles } from '$lib/stores/desktopStore';

const modernStyles = iconStyles.filter(style =>
  styleCategories.modern.includes(style.id)
);
```

---

Last Updated: 2026-01-19
