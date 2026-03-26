<script lang="ts">
  import { onMount } from 'svelte';
  import { Search, ChevronDown, Loader } from 'lucide-svelte';
  import Button from '$lib/ui/button/Button.svelte';
  import Input from '$lib/ui/input/Input.svelte';
  import Card from '$lib/ui/card/Card.svelte';
  import ScrollArea from '$lib/ui/scroll-area/ScrollArea.svelte';
  import { listOntologies, searchClasses, type OntologyInfo } from '$lib/api/ontology';

  let ontologies: OntologyInfo[] = $state([]);
  let selectedOntology: OntologyInfo | null = $state(null);
  let searchQuery = $state('');
  let loading = $state(true);
  let error = $state<string | null>(null);
  let showOntologyDropdown = $state(false);
  let selectedNamespace = $state<string | null>(null);

  const namespaces = [
    'prov:',
    'org:',
    'dcat:',
    'owl:',
    'rdfs:',
    'rdf:',
    'xsd:',
    'foaf:',
    'skos:',
  ];

  onMount(async () => {
    try {
      ontologies = await listOntologies();
      if (ontologies.length > 0) {
        selectedOntology = ontologies[0];
      }
      loading = false;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load ontologies';
      loading = false;
    }
  });

  async function handleSearch() {
    if (!selectedOntology || !searchQuery.trim()) return;

    try {
      const results = await searchClasses(selectedOntology.uri, searchQuery);
      console.log('Search results:', results);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Search failed';
    }
  }

  function selectOntology(onto: OntologyInfo) {
    selectedOntology = onto;
    showOntologyDropdown = false;
    searchQuery = '';
  }

  function getNamespacePrefix(uri: string): string {
    const match = uri.match(/[a-z]+:$/);
    return match ? match[0] : '';
  }
</script>

<div class="w-full h-screen flex flex-col bg-gray-50 dark:bg-gray-900">
  <!-- Header -->
  <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-6">
    <div class="max-w-7xl mx-auto">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">Ontology Explorer</h1>
      <p class="text-gray-600 dark:text-gray-400">
        Browse and explore loaded ontologies, classes, and properties
      </p>
    </div>
  </div>

  {#if loading}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <Loader class="mx-auto mb-4 animate-spin" size={32} />
        <p class="text-gray-600 dark:text-gray-400">Loading ontologies...</p>
      </div>
    </div>
  {:else if error}
    <div class="flex-1 flex items-center justify-center">
      <Card class="max-w-md p-6 bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800">
        <h3 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">Error</h3>
        <p class="text-red-700 dark:text-red-300">{error}</p>
      </Card>
    </div>
  {:else}
    <div class="flex-1 flex overflow-hidden">
      <!-- Sidebar -->
      <div class="w-80 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col">
        <!-- Ontology Selector -->
        <div class="p-4 border-b border-gray-200 dark:border-gray-700">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Select Ontology
          </label>
          <div class="relative">
            <button
              type="button"
              class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-600 flex items-center justify-between"
              onclick={() => (showOntologyDropdown = !showOntologyDropdown)}
            >
              <span class="truncate">{selectedOntology?.name || 'Choose...'}</span>
              <ChevronDown size={16} />
            </button>

            {#if showOntologyDropdown}
              <div class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-lg z-10">
                <ScrollArea class="max-h-64">
                  {#each ontologies as onto (onto.uri)}
                    <button
                      type="button"
                      class="w-full text-left px-4 py-2 hover:bg-blue-50 dark:hover:bg-blue-900 text-gray-900 dark:text-white text-sm {selectedOntology?.uri ===
                      onto.uri
                        ? 'bg-blue-100 dark:bg-blue-900 font-medium'
                        : ''}"
                      onclick={() => selectOntology(onto)}
                    >
                      <div class="font-medium">{onto.name}</div>
                      <div class="text-xs text-gray-500 dark:text-gray-400">{onto.prefix}</div>
                    </button>
                  {/each}
                </ScrollArea>
              </div>
            {/if}
          </div>
        </div>

        <!-- Search Box -->
        <div class="p-4 border-b border-gray-200 dark:border-gray-700">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Search Classes
          </label>
          <div class="flex gap-2">
            <Input
              type="text"
              placeholder="Search by name..."
              bind:value={searchQuery}
              class="flex-1 text-sm"
              onkeydown={(e) => e.key === 'Enter' && handleSearch()}
            />
            <Button
              size="sm"
              onclick={handleSearch}
              class="px-3"
              title="Search"
            >
              <Search size={16} />
            </Button>
          </div>
        </div>

        <!-- Namespace Filter -->
        <div class="p-4 border-b border-gray-200 dark:border-gray-700">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Filter by Namespace
          </label>
          <div class="space-y-2">
            <button
              type="button"
              class={`w-full text-left px-3 py-2 rounded text-sm transition-colors ${
                selectedNamespace === null
                  ? 'bg-blue-100 dark:bg-blue-900 text-blue-900 dark:text-blue-100 font-medium'
                  : 'hover:bg-gray-100 dark:hover:bg-gray-700'
              }`}
              onclick={() => (selectedNamespace = null)}
            >
              All
            </button>
            {#each namespaces as ns}
              <button
                type="button"
                class={`w-full text-left px-3 py-2 rounded text-sm transition-colors ${
                  selectedNamespace === ns
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-900 dark:text-blue-100 font-medium'
                    : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
                onclick={() => (selectedNamespace = ns)}
              >
                {ns}
              </button>
            {/each}
          </div>
        </div>

        <!-- Ontology Stats -->
        {#if selectedOntology}
          <div class="p-4 flex-1 overflow-y-auto">
            <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">Statistics</h3>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-600 dark:text-gray-400">Classes:</span>
                <span class="font-medium text-gray-900 dark:text-white">{selectedOntology.classCount}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-600 dark:text-gray-400">Properties:</span>
                <span class="font-medium text-gray-900 dark:text-white">
                  {selectedOntology.propertyCount}
                </span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-600 dark:text-gray-400">Imports:</span>
                <span class="font-medium text-gray-900 dark:text-white">
                  {selectedOntology.importedOntologies.length}
                </span>
              </div>
            </div>

            {#if selectedOntology.importedOntologies.length > 0}
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mt-4 mb-2">
                Imported
              </h3>
              <div class="space-y-1">
                {#each selectedOntology.importedOntologies as imported}
                  <div class="text-xs text-gray-600 dark:text-gray-400 truncate" title={imported}>
                    {imported.split('/').pop()?.split('#').pop() || imported}
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Main Content -->
      <div class="flex-1 flex flex-col overflow-hidden p-6">
        {#if selectedOntology}
          <div class="flex-1 flex gap-6 overflow-hidden">
            <!-- Classes Panel -->
            <div class="flex-1 flex flex-col bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
              <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  Classes ({selectedOntology.classCount})
                </h2>
              </div>
              <ScrollArea class="flex-1">
                <div class="p-4">
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    Class hierarchy view coming soon. Classes can be explored through the detail pages.
                  </p>
                </div>
              </ScrollArea>
            </div>

            <!-- Properties Panel -->
            <div class="flex-1 flex flex-col bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
              <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  Properties ({selectedOntology.propertyCount})
                </h2>
              </div>
              <ScrollArea class="flex-1">
                <div class="p-4">
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    Property definitions and domains will be displayed here.
                  </p>
                </div>
              </ScrollArea>
            </div>
          </div>
        {:else}
          <div class="flex items-center justify-center h-full">
            <p class="text-gray-500 dark:text-gray-400">No ontologies available</p>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
