import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
	listDeals,
	getDeal,
	createDeal,
	updateDeal,
	deleteDeal,
	verifyCompliance,
	type Deal,
	type CreateDealRequest
} from '$lib/api/deals';

// Mock fetch
global.fetch = vi.fn();

const mockDeal: Deal = {
	id: 'deal-001',
	name: 'Test Deal',
	amount: 100000,
	currency: 'USD',
	status: 'draft',
	buyerId: 'buyer-001',
	sellerId: 'seller-001',
	expectedCloseDate: '2026-04-30',
	probability: 75,
	stage: 'negotiation',
	createdAt: '2026-03-25T10:00:00Z',
	updatedAt: '2026-03-25T10:00:00Z',
	rdfTripleCount: 42,
	complianceStatus: 'pending',
	kycVerified: false,
	amlScreening: 'pending',
	domain: 'Finance'
};

describe('Deals API Client', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('listDeals', () => {
		it('should fetch deals with default pagination', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: [mockDeal] })
			});

			const deals = await listDeals();

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('/api/deals?limit=20&offset=0'),
				expect.any(Object)
			);
			expect(deals).toEqual([mockDeal]);
		});

		it('should apply status filter', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: [] })
			});

			await listDeals(20, 0, 'active');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('status=active'),
				expect.any(Object)
			);
		});

		it('should apply domain filter', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: [] })
			});

			await listDeals(20, 0, undefined, 'Finance');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('domain=Finance'),
				expect.any(Object)
			);
		});
	});

	describe('getDeal', () => {
		it('should fetch a single deal by ID', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: mockDeal })
			});

			const deal = await getDeal('deal-001');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('/api/deals/deal-001'),
				expect.any(Object)
			);
			expect(deal).toEqual(mockDeal);
		});

		it('should throw error if deal not found', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: null })
			});

			await expect(getDeal('nonexistent')).rejects.toThrow('Deal not found');
		});
	});

	describe('createDeal', () => {
		it('should create a new deal', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: mockDeal })
			});

			const request: CreateDealRequest = {
				name: 'Test Deal',
				amount: 100000,
				currency: 'USD',
				buyerId: 'buyer-001',
				sellerId: 'seller-001'
			};

			const deal = await createDeal(request);

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('/api/deals'),
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify(request)
				})
			);
			expect(deal).toEqual(mockDeal);
		});

		it('should throw error if creation fails', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ error: 'Invalid amount' })
			});

			const request: CreateDealRequest = {
				name: 'Test Deal',
				amount: -1000,
				currency: 'USD',
				buyerId: 'buyer-001',
				sellerId: 'seller-001'
			};

			await expect(createDeal(request)).rejects.toThrow('Invalid amount');
		});
	});

	describe('updateDeal', () => {
		it('should update an existing deal', async () => {
			const mockFetch = global.fetch as any;
			const updatedDeal = { ...mockDeal, status: 'active' as const };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: updatedDeal })
			});

			const deal = await updateDeal('deal-001', { status: 'active' });

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('/api/deals/deal-001'),
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ status: 'active' })
				})
			);
			expect(deal.status).toBe('active');
		});
	});

	describe('deleteDeal', () => {
		it('should delete a deal', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({})
			});

			await deleteDeal('deal-001');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('/api/deals/deal-001'),
				expect.objectContaining({
					method: 'DELETE'
				})
			);
		});
	});

	describe('verifyCompliance', () => {
		it('should verify deal compliance', async () => {
			const mockFetch = global.fetch as any;
			const verifiedDeal = { ...mockDeal, complianceStatus: 'pass' as const };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: verifiedDeal })
			});

			const deal = await verifyCompliance('deal-001');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('/api/deals/deal-001/verify-compliance'),
				expect.objectContaining({
					method: 'POST'
				})
			);
			expect(deal.complianceStatus).toBe('pass');
		});
	});

	describe('Error Handling', () => {
		it('should handle HTTP errors', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: false,
				status: 500,
				json: async () => ({ error: 'Server error' })
			});

			await expect(listDeals()).rejects.toThrow('Server error');
		});

		it('should handle timeout', async () => {
			const mockFetch = global.fetch as any;
			const abortError = new Error('Aborted');
			abortError.name = 'AbortError';
			mockFetch.mockRejectedValueOnce(abortError);

			await expect(listDeals()).rejects.toThrow('Request timeout');
		});

		it('should handle network errors', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockRejectedValueOnce(new Error('Network error'));

			await expect(listDeals()).rejects.toThrow('Network error');
		});
	});

	describe('Request Headers', () => {
		it('should include Content-Type header', async () => {
			const mockFetch = global.fetch as any;
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ data: [mockDeal] })
			});

			await listDeals();

			expect(mockFetch).toHaveBeenCalledWith(
				expect.any(String),
				expect.objectContaining({
					headers: expect.objectContaining({
						'Content-Type': 'application/json'
					})
				})
			);
		});
	});
});
