import { describe, it, expect } from 'vitest';
import { getUrls } from './utils';

describe('utils', () => {
	describe('getUrls', () => {
		it('should extract URLs from text', () => {
			const text = 'Check this out: https://example.com and http://test.org';
			const urls = getUrls(text);
			expect(urls).toEqual(['https://example.com', 'http://test.org']);
		});

		it('should return empty array if no URLs found', () => {
			const text = 'Hello world';
			const urls = getUrls(text);
			expect(urls).toEqual([]);
		});

		it('should handle multiple URLs in a row', () => {
			const text = 'https://a.com https://b.com';
			const urls = getUrls(text);
			expect(urls).toEqual(['https://a.com', 'https://b.com']);
		});
	});
});
