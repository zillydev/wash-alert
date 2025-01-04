// router/index.js
import { createRouter, createWebHistory } from 'vue-router';
import Home from '../pages/Home.vue';
import TestPage from '../pages/TestPage.vue';

// Define the routes
const routes = [
  {
    path: '/',
    component: Home
  },
  {
    path: '/test',
    component: TestPage
  }
];

// Create the router instance
const router = createRouter({
  history: createWebHistory(), // Use HTML5 history mode for cleaner URLs
  routes
});

export default router;
