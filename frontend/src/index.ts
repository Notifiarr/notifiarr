import { mount } from 'svelte'
import '/src/assets/app.css'
import Index from '/src/Index.svelte'
import '/src/includes/locale/index.svelte.ts'

export default mount(Index, { target: document.getElementById('notifiarr')! })
