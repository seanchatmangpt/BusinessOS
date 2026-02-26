<!--
  ThemeToggle.svelte
  Toggle between light and dark themes
  Converted from Next.js theme-toggle.tsx
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import Button from './Button.svelte';
  import SunIcon from './icons/SunIcon.svelte';
  import MoonIcon from './icons/MoonIcon.svelte';
  import { themeStore } from '$lib/stores/themeStore';

  let theme: 'light' | 'dark' = 'light';

  // Subscribe to theme store
  const unsubscribe = themeStore.subscribe(state => {
    theme = state.resolvedTheme;
  });

  onMount(() => {
    return () => unsubscribe();
  });

  function toggleTheme() {
    const newTheme = theme === 'dark' ? 'light' : 'dark';
    themeStore.setTheme(newTheme);
  }
</script>

<Button
  variant="ghost"
  size="icon"
  on:click={toggleTheme}
  className="theme-toggle rounded-full"
>
  {#if theme === 'dark'}
    <SunIcon size={20} />
  {:else}
    <MoonIcon size={20} />
  {/if}
  <span class="sr-only">Toggle theme</span>
</Button>

<style>
  :global(.theme-toggle) {
    width: 2.75rem !important;
    height: 2.75rem !important;
    border-radius: 9999px !important;
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border-width: 0;
  }
</style>
