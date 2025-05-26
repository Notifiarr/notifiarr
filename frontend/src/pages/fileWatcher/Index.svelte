<script lang="ts" module>
  import { faEyeEvil, faFileWaveform } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'FileWatcher',
    i: faFileWaveform,
    d1: 'thistle',
    d2: 'blue',
    c1: 'sienna',
    c2: 'moccasin',
  }
</script>

<script lang="ts">
  import { CardBody } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Watcher from './Watcher.svelte'
  import { profile } from '../../api/profile.svelte'
  import type { App } from '../../includes/Instance.svelte'
  import { deepCopy } from '../../includes/util'
  import type { Config, WatchFile } from '../../api/notifiarrConfig'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import Instances from '../../includes/Instances.svelte'

  const app: App = {
    id: 'FileWatcher',
    name: 'File Watcher',
    logo: faEyeEvil,
    iconProps: { c1: 'wheat', c2: 'firebrick', d2: 'purple' },
    disabled: [],
    hidden: [],
    empty: {
      path: '',
      regex: '',
      skip: '',
      poll: false,
      pipe: false,
      mustExist: false,
      logMatch: false,
    } as WatchFile,
    merge: (index: number, form: WatchFile): Config => {
      const c = deepCopy($profile.config)
      if (!c.watchFiles) c.watchFiles = []
      for (let i = 0; i < c.watchFiles.length; i++) {
        if (i === index) c.watchFiles[i] = form
        else c.watchFiles[i] = {} as WatchFile
      }
      return c
    },
  }
  // Handle form submission
  function submit(e: Event) {
    e.preventDefault()
    // profile.writeConfig(c)
  }
  const flt = new FormListTracker($profile.config.watchFiles, app)
</script>

<Header {page} />
<CardBody>
  <Instances {flt} Child={Watcher}>
    {#snippet headerActive(index)}
      {index + 1}. {flt.original[index]?.path}
    {/snippet}
  </Instances>
</CardBody>
<Footer {submit} />
