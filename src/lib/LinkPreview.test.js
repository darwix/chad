import { render, waitFor } from '@testing-library/svelte';
import { expect, test, vi } from 'vitest';
import LinkPreview from './LinkPreview.svelte';
import { getLinkPreview } from 'link-preview-js';

// Mock link-preview-js
vi.mock('link-preview-js', () => ({
	getLinkPreview: vi.fn()
}));

test('falls back to Microlink API if link-preview-js detects bot block', async () => {
	getLinkPreview.mockResolvedValueOnce({
		title: 'Attention Required! | Cloudflare',
		description: 'Checking your browser before accessing archiveofourown.org.',
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

	const { getByText } = render(LinkPreview, { url: 'https://archiveofourown.org' });

	await waitFor(() => expect(getByText('Fallback Title')).toBeTruthy());
	expect(getByText('Fallback Description')).toBeTruthy();
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
