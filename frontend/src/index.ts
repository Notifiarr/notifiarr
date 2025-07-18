import { mount } from 'svelte'
import Index from '/src/Index.svelte'
import '/src/assets/app.css'
import '/src/includes/locale/index.svelte.ts'
import 'bootstrap/dist/css/bootstrap.min.css'
import '/src/navigation/goto-element.svelte'

export default mount(Index, { target: document.getElementById('app')! })
