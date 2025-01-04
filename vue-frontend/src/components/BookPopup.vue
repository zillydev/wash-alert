<script setup lang="ts">
// Props to control visibility of the popup
defineProps({
  isVisible: {
    type: Boolean,
    default: false,
  },
});

// Emits to update the visibility in the parent component
const emit = defineEmits();

// Method to close the popup
const closePopup = () => {
  emit('update:isVisible', false);
};
</script>

<template>
  <div v-if="isVisible" class="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center" @click="closePopup">
    <div class="bg-white dark:bg-gray-800 p-4 rounded-md w-[80%]" @click.stop>
      <!-- <button @click="closePopup">Close</button> -->
      <slot></slot>
    </div>
  </div>
</template>
