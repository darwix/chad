<script>
  import { onMount } from 'svelte';

  let { username } = $props();

  let messages = $state([]);
  let newMessage = $state('');
  let socket;
  let onlineUsers = $state([]);
  let typingUsers = $state(new Set());
  let typingTimeout;
  let chatContainer;

  const emojis = ['😀', '😂', '😍', '👍', '🔥', '🎉', '❤️', '🤔', '🙌', '✨'];

  function connect() {
    socket = new WebSocket('ws://localhost:8080/ws');

    socket.onopen = () => {
      console.log('Connected to server');
      // Send initial presence
      socket.send(JSON.stringify({
        type: 'presence',
        sender: username
      }));
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'chat') {
        messages = [...messages, data];
      } else if (data.type === 'presence') {
        onlineUsers = data.users;
      } else if (data.type === 'typing') {
        if (data.sender !== username) {
          if (data.is_typing) {
            typingUsers.add(data.sender);
          } else {
            typingUsers.delete(data.sender);
          }
          typingUsers = new Set(typingUsers); // Trigger reactivity
        }
      }
    };

    socket.onclose = () => {
      console.log('Disconnected, retrying in 3s...');
      setTimeout(connect, 3000);
    };
  }

  async function fetchHistory() {
    try {
      const res = await fetch('http://localhost:8080/history');
      if (res.ok) {
        messages = await res.json();
      }
    } catch (e) {
      console.error('Failed to fetch history', e);
    }
  }

  onMount(() => {
    fetchHistory();
    connect();
    return () => socket?.close();
  });

  $effect(() => {
    // This effect runs whenever messages change
    messages;
    if (chatContainer) {
      chatContainer.scrollTop = chatContainer.scrollHeight;
    }
  });

  function sendMessage() {
    if (newMessage.trim() && socket) {
      const msg = {
        type: 'chat',
        sender: username,
        content: newMessage
      };
      socket.send(JSON.stringify(msg));
      newMessage = '';
      sendTyping(false);
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
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: 'typing',
        sender: username,
        is_typing: isTyping
      }));
    }
  }

  function addEmoji(emoji) {
    newMessage += emoji;
  }

  function formatTime(ts) {
    return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
</script>

<div class="flex h-screen bg-gray-100">
  <!-- Sidebar -->
  <div class="w-64 bg-white border-r flex flex-col">
    <div class="p-4 border-b font-bold text-lg">Online Users ({onlineUsers.length})</div>
    <div class="flex-1 overflow-y-auto p-4 space-y-2">
      {#each onlineUsers as user (user)}
        <div class="flex items-center space-x-2">
          <div class="w-2 h-2 bg-green-500 rounded-full"></div>
          <span>{user} {user === username ? '(You)' : ''}</span>
        </div>
      {/each}
    </div>
  </div>

  <!-- Main Chat -->
  <div class="flex-1 flex flex-col">
    <!-- Messages -->
    <div
      bind:this={chatContainer}
      class="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50"
    >
      {#each messages as msg (msg.id || msg.timestamp)}
        <div class="flex flex-col {msg.sender === username ? 'items-end' : 'items-start'}">
          <div class="flex items-baseline space-x-2">
            <span class="text-xs font-semibold text-gray-500">{msg.sender}</span>
            <span class="text-[10px] text-gray-400">{formatTime(msg.timestamp)}</span>
          </div>
          <div class="mt-1 px-4 py-2 rounded-lg max-w-md {msg.sender === username ? 'bg-blue-600 text-white' : 'bg-white text-gray-800 shadow-sm'}">
            {msg.content}
          </div>
        </div>
      {/each}

      {#if typingUsers.size > 0}
        <div class="text-xs text-gray-400 italic">
          {Array.from(typingUsers).join(', ')} {typingUsers.size === 1 ? 'is' : 'are'} typing...
        </div>
      {/if}
    </div>

    <!-- Input -->
    <div class="p-4 bg-white border-t">
      <div class="flex flex-wrap gap-2 mb-2">
        {#each emojis as emoji}
          <button
            onclick={() => addEmoji(emoji)}
            class="hover:bg-gray-100 p-1 rounded"
          >
            {emoji}
          </button>
        {/each}
      </div>
      <div class="flex space-x-2">
        <input
          type="text"
          bind:value={newMessage}
          onkeydown={handleKeydown}
          placeholder="Type a message..."
          class="flex-1 border border-gray-300 rounded-full px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          onclick={sendMessage}
          class="bg-blue-600 text-white rounded-full px-6 py-2 hover:bg-blue-700 focus:outline-none"
        >
          Send
        </button>
      </div>
    </div>
  </div>
</div>
