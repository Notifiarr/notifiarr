import { mount } from 'svelte'
import './app.css'
import Index from './Index.svelte'

export default mount(Index, {
  target: document.getElementById('notifiarr')!,
})

