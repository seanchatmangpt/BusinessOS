<!--
  SilverOrb.svelte
  The OSA silver orb branding element
-->
<script lang="ts">
  export let size: number = 80;
  export let className: string = "";
</script>

<div
  class="orb-container {className}"
  style="width: {size}px; height: {size}px;"
>
  <div class="orb-body animate-pulse-orb">
    <div class="orb-shimmer animate-shimmer"></div>
  </div>
</div>

<style>
  .orb-container {
    position: relative;
  }

  .orb-body {
    position: absolute;
    inset: 0;
    border-radius: 9999px;
    background: linear-gradient(
      135deg,
      #f8f9fa 0%,
      #d1d5db 25%,
      #9ca3af 50%,
      #6b7280 75%,
      #4b5563 100%
    );
    box-shadow:
      0 0 40px rgba(156, 163, 175, 0.4),
      inset 0 0 20px rgba(255, 255, 255, 0.5);
  }

  .orb-shimmer {
    position: absolute;
    inset: 0;
    border-radius: 9999px;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.6),
      transparent
    );
    background-size: 1000px 100%;
  }

  @keyframes pulse-orb {
    0%, 100% {
      transform: scale(1);
      filter: brightness(1);
    }
    50% {
      transform: scale(1.05);
      filter: brightness(1.15);
    }
  }

  @keyframes shimmer {
    0% {
      background-position: -1000px 0;
    }
    100% {
      background-position: 1000px 0;
    }
  }

  .animate-pulse-orb {
    animation: pulse-orb 2s ease-in-out infinite;
  }

  .animate-shimmer {
    animation: shimmer 3s linear infinite;
  }
</style>
