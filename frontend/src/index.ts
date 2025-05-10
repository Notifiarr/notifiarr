import { mount } from 'svelte'
import '/src/app.css'
import Index from '/src/Index.svelte'
import '/src/lib/locale'

export default mount(Index, { target: document.getElementById('notifiarr')! })
