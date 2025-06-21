import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'

// Definizione delle rotte principali dell'applicazione Vue.
// Ogni rotta collega un percorso a una vista specifica.
// Le rotte '/' e '/login' portano alla schermata di login, '/home' alla home principale.
const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{ path: '/', component: LoginView },
		{ path: '/login', component: LoginView },
		{ path: '/home', component: HomeView },
	]
})

export default router
