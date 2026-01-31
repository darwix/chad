import { render, waitFor } from '@testing-library/svelte';
import { expect, test, vi } from 'vitest';
import LinkPreview from './LinkPreview.svelte';
import { getLinkPreview } from 'link-preview-js';

// Mock link-preview-js
vi.mock('link-preview-js', () => ({
	getLinkPreview: vi.fn()
}));

test('uses Microlink API immediately for difficult domains like AO3', async () => {
	// Mock fetch for Microlink
	global.fetch = vi.fn().mockResolvedValueOnce({
		json: () =>
			Promise.resolve({
				status: 'success',
				data: {
					title: 'AO3 Title',
					description: 'AO3 Description',
					image: { url: 'https://archiveofourown.org/logo.png' }
				}
			})
	});

	const { getByText } = render(LinkPreview, { url: 'https://archiveofourown.org/tags/test' });

	await waitFor(() => expect(getByText('AO3 Title')).toBeTruthy());
	expect(getByText('AO3 Description')).toBeTruthy();
	expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('api.microlink.io'));
	// getLinkPreview should NOT have been called yet for difficult domains
	expect(getLinkPreview).not.toHaveBeenCalled();
});

test('falls back to Microlink API if link-preview-js detects bot block on unknown domain', async () => {
	getLinkPreview.mockResolvedValueOnce({
		title: 'Attention Required! | Cloudflare',
		description: 'Checking your browser...',
		images: []
	});

	// Mock fetch for the fallback
	global.fetch = vi.fn().mockResolvedValueOnce({
		json: () =>
			Promise.resolve({
				status: 'success',
				data: {
					title: 'Fallback Title',
					description: 'Fallback Description',
					image: { url: 'https://fallback.com/image.png' }
				}
			})
	});

	const { getByText } = render(LinkPreview, { url: 'https://unknown-blocked-site.com' });

	await waitFor(() => expect(getByText('Fallback Title')).toBeTruthy());
	expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('api.microlink.io'));
});

test('renders correctly for valid metadata', async () => {
	getLinkPreview.mockResolvedValueOnce({
		title: 'Google',
		description: 'Search the world information',
		images: ['https://google.com/logo.png']
	});

	const { getByText } = render(LinkPreview, { url: 'https://google.com' });

	await waitFor(() => expect(getByText('Google')).toBeTruthy());
	expect(getByText('Search the world information')).toBeTruthy();
});
