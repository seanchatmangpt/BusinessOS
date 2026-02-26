<script lang="ts">
	import type { ModuleManifest } from '$lib/types/modules';

	interface Props {
		manifest: ModuleManifest;
	}

	let { manifest }: Props = $props();

	const jsonString = $derived(JSON.stringify(manifest, null, 2));
</script>

<div class="bg-gray-900 rounded-xl p-4 overflow-x-auto">
	<pre class="text-sm text-gray-100 font-mono"><code>{jsonString}</code></pre>
</div>
