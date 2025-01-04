<script setup lang="ts">
import { ref, inject } from 'vue'

const message = ref('')

const socket = inject('socket') as WebSocket
socket.onmessage = (event) => {
  console.log(JSON.parse(event.data))
  message.value = JSON.parse(event.data)[0]["FloorNumber"]
}
</script>

<template>
  <div class="container flex justify-center items-center">
    <h1 class="text-3xl font-bold underline bg-red-500">
      {{ message.valueOf() }}
    </h1>
  </div>
</template>
