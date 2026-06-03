import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'

// Define the route mapping configuration
const routes = [
  {
    path: '/login',
    component: LoginView,
    meta: {requiresAuth: false} // Public page, accessible by anyone
  },
  { 
    path: '/dashboard', 
    component: DashboardView,
    meta: { requiresAuth: true } // Protected page, requires valid authentication token
  },
  { 
    path: '/:pathMatch(.*)*', 
    redirect: '/login'           // Catch-all route to redirect any invalid URL back to the login page
  }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Global Navigation Guard
router.beforeEach((to, from, next) => {
  // Retrieve the security token provided by AWS from the browser's local storage
  const token = localStorage.getItem('id_token')

  // Intercept if the destination requires authentication but no token is present
  if (to.meta.requiresAuth && !token) {
    next('/login') // Block access and redirect back to the entry portal
  } else {
    next() // Grant passage if the token exists or if accessing a public page
  }
})

export default router
