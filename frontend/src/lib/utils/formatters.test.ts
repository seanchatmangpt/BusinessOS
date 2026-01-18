import { describe, it, expect } from 'vitest';
import { getInitials, formatDate, formatCurrency } from './formatters';

describe('formatters', () => {
	describe('getInitials', () => {
		it('returns initials for full name', () => {
			expect(getInitials('John Doe')).toBe('JD');
		});

		it('returns single initial for one word', () => {
			expect(getInitials('John')).toBe('J');
		});

		it('returns "?" for empty string', () => {
			expect(getInitials('')).toBe('?');
		});

		it('handles names with extra spaces', () => {
			expect(getInitials('  John   Doe  ')).toBe('JD');
		});

		it('returns first and last initials for multi-word names', () => {
			expect(getInitials('John Michael Doe')).toBe('JD');
		});

		it('returns uppercase initials', () => {
			expect(getInitials('john doe')).toBe('JD');
		});
	});

	describe('formatDate', () => {
		it('formats date in short format by default', () => {
			const date = new Date('2024-01-15');
			const result = formatDate(date, 'short');
			expect(result).toMatch(/01\/15\/2024/);
		});

		it('formats date in long format', () => {
			const date = new Date('2024-01-15');
			const result = formatDate(date, 'long');
			expect(result).toContain('January');
			expect(result).toContain('15');
			expect(result).toContain('2024');
		});

		it('formats date string input', () => {
			const result = formatDate('2024-01-15', 'short');
			expect(result).toMatch(/01\/15\/2024/);
		});

		it('returns "Today" for relative format on current date', () => {
			const today = new Date();
			const result = formatDate(today, 'relative');
			expect(result).toBe('Today');
		});

		it('returns "Yesterday" for relative format on yesterday', () => {
			const yesterday = new Date();
			yesterday.setDate(yesterday.getDate() - 1);
			const result = formatDate(yesterday, 'relative');
			expect(result).toBe('Yesterday');
		});

		it('returns days ago for dates within a week', () => {
			const threeDaysAgo = new Date();
			threeDaysAgo.setDate(threeDaysAgo.getDate() - 3);
			const result = formatDate(threeDaysAgo, 'relative');
			expect(result).toBe('3 days ago');
		});
	});

	describe('formatCurrency', () => {
		it('formats currency with default USD', () => {
			expect(formatCurrency(1234.56)).toBe('$1,234.56');
		});

		it('formats currency with no decimals for whole numbers', () => {
			expect(formatCurrency(1234)).toContain('$1,234');
		});

		it('formats large numbers with thousand separators', () => {
			const result = formatCurrency(1234567.89);
			expect(result).toContain('1,234,567');
		});

		it('formats zero correctly', () => {
			expect(formatCurrency(0)).toContain('$0');
		});

		it('formats negative numbers', () => {
			const result = formatCurrency(-100);
			expect(result).toContain('-');
			expect(result).toContain('100');
		});

		it('formats with different currency codes', () => {
			const resultEUR = formatCurrency(100, 'EUR');
			expect(resultEUR).toContain('100');
			// Note: actual symbol depends on locale, but amount should be there
		});
	});
});
