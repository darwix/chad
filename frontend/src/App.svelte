<script>
  import { onMount } from 'svelte';
  import Login from './lib/Login.svelte';
  import Chat from './lib/Chat.svelte';

  let username = $state(null);
  let loading = $state(true);

  onMount(() => {
    const storedName = localStorage.getItem('chat_username');
    if (storedName) {
      username = storedName;
    }
    loading = false;
  });

  function handleLogin(name) {
    username = name;
  }

  function handleLogout() {
    localStorage.removeItem('chat_username');
    username = null;
  }
</script>

<main class="h-screen w-full bg-gray-100">
  {#if loading}
    <div class="flex items-center justify-center h-full">
      <p>Loading...</p>
    </div>
  {:else if !username}
    <Login onlogin={handleLogin} />
  {:else}
    <div class="absolute top-4 right-4 z-10">
      <button
        onclick={handleLogout}
        class="text-xs text-gray-500 hover:text-red-500 bg-white px-2 py-1 rounded shadow-sm"
      >
        Change Name
      </button>
    </div>
    <Chat {username} />
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  }
</style>
