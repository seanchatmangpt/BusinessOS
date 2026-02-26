<script lang="ts">
	import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-svelte';

	interface Props {
		page: number;
		pageSize: number;
		total: number;
		onPageChange?: (newPage: number) => void;
	}

	let { page, pageSize, total, onPageChange }: Props = $props();

	const totalPages = $derived(Math.ceil(total / pageSize));
	const startItem = $derived((page - 1) * pageSize + 1);
	const endItem = $derived(Math.min(page * pageSize, total));

	const isFirstPage = $derived(page === 1);
	const isLastPage = $derived(page === totalPages);

	function goToPage(newPage: number) {
		if (newPage >= 1 && newPage <= totalPages && onPageChange) {
			onPageChange(newPage);
		}
	}

	function goToFirstPage() {
		goToPage(1);
	}

	function goToPreviousPage() {
		goToPage(page - 1);
	}

	function goToNextPage() {
		goToPage(page + 1);
	}

	function goToLastPage() {
		goToPage(totalPages);
	}

	// Determine which page numbers to show
	const pageNumbers = $derived.by(() => {
		const pages: (number | string)[] = [];
		const maxPages = 7;
		const siblingCount = 1;

		if (totalPages <= maxPages) {
			// Show all pages if 7 or fewer
			for (let i = 1; i <= totalPages; i++) {
				pages.push(i);
			}
		} else {
			// Always show first page
			pages.push(1);

			// Calculate left and right range around current page
			const leftStart = Math.max(2, page - siblingCount);
			const rightEnd = Math.min(totalPages - 1, page + siblingCount);

			// Add left ellipsis if needed
			if (leftStart > 2) {
				pages.push('...');
			}

			// Add pages around current page
			for (let i = leftStart; i <= rightEnd; i++) {
				pages.push(i);
			}

			// Add right ellipsis if needed
			if (rightEnd < totalPages - 1) {
				pages.push('...');
			}

			// Always show last page
			pages.push(totalPages);
		}

		return pages;
	});
</script>

{#if total > 0}
	<div class="flex flex-col gap-4 items-center justify-between py-4 px-2 sm:flex-row sm:px-0 mt-6 pt-4 border-t border-gray-200">
		<!-- Results info text -->
		<div class="text-sm text-muted-foreground">
			<span>Showing <span class="font-medium">{startItem}</span> to <span class="font-medium">{endItem}</span> of <span class="font-medium">{total}</span> results</span>
		</div>

		<!-- Pagination controls -->
		<div class="flex items-center gap-1 sm:gap-2">
			<!-- First page button -->
			<button
				type="button"
				onclick={goToFirstPage}
				disabled={isFirstPage}
				class="inline-flex items-center justify-center h-9 w-9 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50 transition-colors"
				title="Go to first page"
				aria-label="First page"
			>
				<ChevronsLeft class="h-4 w-4" />
			</button>

			<!-- Previous page button -->
			<button
				type="button"
				onclick={goToPreviousPage}
				disabled={isFirstPage}
				class="inline-flex items-center justify-center h-9 w-9 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50 transition-colors"
				title="Go to previous page"
				aria-label="Previous page"
			>
				<ChevronLeft class="h-4 w-4" />
			</button>

			<!-- Page numbers -->
			<div class="flex items-center gap-1">
				{#each pageNumbers as pageNum, idx (idx)}
					{#if pageNum === '...'}
						<span class="text-muted-foreground px-2">•••</span>
					{:else}
						<button
							type="button"
							onclick={() => goToPage(pageNum as number)}
							class="inline-flex items-center justify-center h-9 w-9 rounded-md border text-sm font-medium transition-colors {page === pageNum
								? 'border-primary bg-primary text-primary-foreground'
								: 'border-input bg-background hover:bg-accent hover:text-accent-foreground'}"
							aria-label="Go to page {pageNum}"
							aria-current={page === pageNum ? 'page' : undefined}
						>
							{pageNum}
						</button>
					{/if}
				{/each}
			</div>

			<!-- Next page button -->
			<button
				type="button"
				onclick={goToNextPage}
				disabled={isLastPage}
				class="inline-flex items-center justify-center h-9 w-9 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50 transition-colors"
				title="Go to next page"
				aria-label="Next page"
			>
				<ChevronRight class="h-4 w-4" />
			</button>

			<!-- Last page button -->
			<button
				type="button"
				onclick={goToLastPage}
				disabled={isLastPage}
				class="inline-flex items-center justify-center h-9 w-9 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50 transition-colors"
				title="Go to last page"
				aria-label="Last page"
			>
				<ChevronsRight class="h-4 w-4" />
			</button>
		</div>
	</div>
{/if}
