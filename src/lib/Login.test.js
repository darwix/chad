import { render, screen } from '@testing-library/svelte';
import { expect, test, vi } from 'vitest';
import Login from './Login.svelte';

// Mock Supabase
vi.mock('./supabase', () => ({
	supabase: {
		auth: {
			getSession: vi.fn(() => Promise.resolve({ data: { session: null } })),
			onAuthStateChange: vi.fn(() => ({
				data: { subscription: { unsubscribe: vi.fn() } }
			})),
			signInWithPassword: vi.fn()
		}
	}
}));

test('renders login header', () => {
	render(Login);
	const header = screen.getByRole('heading', { name: 'Login' });
	expect(header).toBeTruthy();
});
