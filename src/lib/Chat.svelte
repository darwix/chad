<script>
	import { onMount } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import LinkPreview from './LinkPreview.svelte';
	import { supabase } from './supabase';
	import { getUrls, formatTime } from './utils';

	let { username } = $props();

	let messages = $state([]);
	let newMessage = $state('');
	let onlineUsers = $state([]);
	let typingUsers = new SvelteSet();
	let typingTimeout;
	let chatContainer;
	let channel;
	let isSidebarOpen = $state(false);
	let isEmojiPickerOpen = $state(false);

	async function fetchHistory() {
		const { data, error } = await supabase
			.from('messages')
			.select('*')
			.order('created_at', { ascending: false })
			.limit(50);

		if (error) {
			console.error('Error fetching history:', error);
		} else {
			messages = data.reverse();
		}
	}

	function setupSupabase() {
		channel = supabase.channel('global-chat', {
			config: {
				presence: {
					key: username
				}
			}
		});

		// Listen for new messages via Postgres Changes
		channel
			.on(
				'postgres_changes',
				{ event: 'INSERT', schema: 'public', table: 'messages' },
				(payload) => {
					messages = [...messages, payload.new];
				}
			)
			// Listen for typing indicators via Broadcast
			.on('broadcast', { event: 'typing' }, (payload) => {
				if (payload.payload.sender !== username) {
					if (payload.payload.is_typing) {
						typingUsers.add(payload.payload.sender);
					} else {
						typingUsers.delete(payload.payload.sender);
					}
				}
			})
			// Track Presence (Online Users)
			.on('presence', { event: 'sync' }, () => {
				const state = channel.presenceState();
				onlineUsers = Object.keys(state);
			})
			.subscribe(async (status) => {
				if (status === 'SUBSCRIBED') {
					await channel.track({ online_at: new Date().toISOString() });
				}
			});
	}

	onMount(() => {
		fetchHistory();
		setupSupabase();
		import('emoji-picker-element');
		return () => {
			if (channel) supabase.removeChannel(channel);
		};
	});

	$effect(() => {
		// Auto-scroll when messages change
		messages;
		if (chatContainer) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
		}
	});

	async function sendMessage() {
		if (newMessage.trim()) {
			const msg = {
				sender: username,
				content: newMessage
			};

			const { error } = await supabase.from('messages').insert([msg]);
			if (error) {
				console.error('Error sending message:', error);
			} else {
				newMessage = '';
				sendTyping(false);
			}
		}
	}

	function handleKeydown(e) {
		if (e.key === 'Enter') {
			sendMessage();
		} else {
			sendTyping(true);
			clearTimeout(typingTimeout);
			typingTimeout = setTimeout(() => sendTyping(false), 2000);
		}
	}

	function sendTyping(isTyping) {
		if (channel) {
			channel.send({
				type: 'broadcast',
				event: 'typing',
				payload: { sender: username, is_typing: isTyping }
			});
		}
	}

	function addEmoji(emoji) {
		newMessage += emoji;
	}

	function toggleSidebar() {
		isSidebarOpen = !isSidebarOpen;
	}

	function toggleEmojiPicker() {
		isEmojiPickerOpen = !isEmojiPickerOpen;
	}
</script>

<div class="flex h-screen bg-gray-100 dark:bg-gray-900 overflow-hidden relative transition-colors duration-200">
	<!-- Sidebar Overlay -->
	{#if isSidebarOpen}
		<button
			class="fixed inset-0 bg-black/50 z-20 md:hidden"
			onclick={toggleSidebar}
			aria-label="Close sidebar"
		></button>
	{/if}

	<!-- Sidebar -->
	<aside
		class="absolute md:relative w-64 h-full bg-white dark:bg-gray-800 border-r dark:border-gray-700 flex flex-col z-30 transition-transform duration-300 transform {isSidebarOpen
			? 'translate-x-0'
			: '-translate-x-full md:translate-x-0'}"
	>
		<div class="p-4 border-b dark:border-gray-700 font-bold text-lg flex justify-between items-center dark:text-white">
			<span>Online Users ({onlineUsers.length})</span>
			<button class="md:hidden text-gray-500 dark:text-gray-400" onclick={toggleSidebar}>
				✕
			</button>
		</div>
		<div class="flex-1 overflow-y-auto p-4 space-y-2">
			{#each onlineUsers as user (user)}
				<div class="flex items-center space-x-2">
					<div class="w-2 h-2 bg-green-500 rounded-full"></div>
					<span class="truncate dark:text-gray-300">{user} {user === username ? '(You)' : ''}</span>
				</div>
			{/each}
		</div>
	</aside>

	<!-- Main Chat -->
	<div class="flex-1 flex flex-col min-w-0">
		<!-- Header -->
		<header class="p-4 bg-white dark:bg-gray-800 border-b dark:border-gray-700 flex items-center md:hidden">
			<button
				class="p-2 -ml-2 text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
				onclick={toggleSidebar}
				aria-label="Toggle sidebar"
			>
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<h1 class="ml-2 font-semibold dark:text-white">Chat</h1>
		</header>

		<!-- Messages -->
		<div bind:this={chatContainer} class="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50 dark:bg-gray-900">
			{#each messages as msg (msg.id || msg.created_at)}
				<div class="flex flex-col {msg.sender === username ? 'items-end' : 'items-start'}">
					<div class="flex items-baseline space-x-2">
						<span class="text-xs font-semibold text-gray-500 dark:text-gray-400">{msg.sender}</span>
						<span class="text-[10px] text-gray-400 dark:text-gray-500">{formatTime(msg.created_at)}</span>
					</div>
					<div
						class="mt-1 px-4 py-2 rounded-lg max-w-[85%] md:max-w-md break-words {msg.sender === username
							? 'bg-blue-600 text-white'
							: 'bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-100 shadow-sm'}"
					>
						{msg.content}
					</div>
					{#each getUrls(msg.content) as url, i (`${url}-${i}`)}
						<LinkPreview {url} />
					{/each}
				</div>
			{/each}

			{#if typingUsers.size > 0}
				<div class="text-xs text-gray-400 dark:text-gray-500 italic px-2">
					{Array.from(typingUsers).join(', ')}
					{typingUsers.size === 1 ? 'is' : 'are'} typing...
				</div>
			{/if}
		</div>

		<!-- Input -->
		<footer class="p-4 bg-white dark:bg-gray-800 border-t dark:border-gray-700 relative">
			{#if isEmojiPickerOpen}
				<div class="absolute bottom-full mb-2 left-4 z-50 shadow-xl rounded-lg overflow-hidden border dark:border-gray-700">
					<emoji-picker
						class={document.documentElement.classList.contains('dark') ? 'dark' : 'light'}
						onemoji-click={(e) => {
							addEmoji(e.detail.unicode);
							isEmojiPickerOpen = false;
						}}
					></emoji-picker>
				</div>
			{/if}
			<div class="flex space-x-2 items-center">
				<button
					onclick={toggleEmojiPicker}
					class="p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 flex items-center justify-center shrink-0"
					aria-label="Add emoji"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
				</button>
				<input
					type="text"
					bind:value={newMessage}
					onkeydown={handleKeydown}
					placeholder="Type a message..."
					class="flex-1 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-full px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
				/>
				<button
					onclick={sendMessage}
					class="bg-blue-600 text-white rounded-full px-5 md:px-6 py-2 hover:bg-blue-700 focus:outline-none text-sm font-medium"
				>
					Send
				</button>
			</div>
		</footer>
	</div>
</div>
