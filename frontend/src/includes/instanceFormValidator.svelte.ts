import { get } from 'svelte/store'
import type { App, Form, Forms } from './Instance.svelte'
import { _ } from './Translate.svelte'
import { deepCopy, deepEqual, delay, success } from './util'

/**
 * InstanceFormValidator is a class that validates the instances in the tabs.
 * it keeps track of the original list of instances, the form-bound list of instances,
 * the removed instances, the invalid instances, and the feedback for the instances.
 * @param instances - The form-bound list of instances in our tabs.
 * @param app - The app we're validating.
 */
export class InstanceFormValidator {
  /** The form-bound list of instances in our tabs. */
  public instances: Forms
  /** The original list of instances in our tabs. */
  public original: Forms
  /** Data about the app we're validating. */
  public app: App
  /** List of removed instances. */
  public removed: Forms = $state([])
  /** List of invalid instances. */
  private feedback: Record<number, Record<string, string>> = $state({})
  /** If any instance in the list has non-empty feedback the form is invalid. */
  public invalid: boolean = $derived(
    Object.values(this.feedback).some(v => Object.values(v).some(v => !!v)),
  )
  /** If the form has changed from the original values. */
  public formChanged: boolean
  /** The active instance tab. */
  public active: number | undefined = $state(0)

  constructor(instances: Forms | undefined, app: App) {
    this.instances = $state(deepCopy(instances ?? []))
    this.original = $state(deepCopy(instances ?? []))
    this.app = app
    this.formChanged = $derived(!deepEqual(this.instances, this.original))
  }

  /** Add a new instance to the list. */
  public addInstance = () => {
    this.instances.push(this.app.empty)
    this.active = this.instances.length - 1
  }

  /** Remove an instance from the list. */
  public delInstance = async (index: number) => {
    // Close the accordion.
    this.active = undefined
    // Wait for it to slide shut.
    await delay(400)
    // Remove the instance from the form (delete the accordion).
    this.instances.splice(index, 1)
    // Reset the feedback for the instance.
    this.feedback = {}
    if (index < this.original.length) {
      this.removed.push(...this.original.splice(index, 1))
      await delay(100)
      success(get(_)('phrases.ItsGone'))
    }
    this.active = 0
  }

  /** Reset the form to the original values. Call this after a form has been submitted. */
  public resetAll = () => {
    this.instances = deepCopy(this.original)
    this.removed = []
    this.validateAll()
  }

  /** Reset a single instance to the original values. Call this when reset button is clicked. */
  public resetForm = (index: number) => {
    this.instances[index] = deepCopy(this.original[index] ?? this.app.empty)
    Object.keys(this.instances[index] ?? {}).forEach(k => {
      this.validate(k, this.instances[index]?.[k as keyof Form], index)
    })
  }

  /** Validate all instances. Call this after a form has been submitted to re-validate any backend changes. */
  private validateAll = () => {
    this.instances.forEach((m, i) => {
      Object.keys(m ?? {}).forEach(k => {
        this.validate(k, m?.[k as keyof Form], i)
      })
    })
  }

  /** Check if an instance is valid.
   * @param index - The index of the current instance the instances list. (0)
   */
  public isValid = (index: number): boolean => {
    return Object.values(this.feedback[index] ?? {}).every(v => !v)
  }

  /** Standard form validator for an integrated instance (plex, sonarr, etc)
   * @param id - The id of the form field. (anything.here.url)
   * @param value - The value of the form field. (http://localhost:8080)
   * @param index - The index of the current instance the instances list. (0)
   * @updates The feedback for the instance.
   */
  public validate = (id: string, value: any, index: number): string => {
    const key = id.split('.').pop()

    if (!this.feedback[index]) this.feedback[index] = {}
    this.feedback[index][id] = ''

    if (key == 'name') {
      this.instances?.forEach((m, i) => {
        if (i !== index && m?.name === value) {
          this.feedback[index][id] = get(_)('phrases.NameInUseByInstance', {
            values: { number: i + 1 },
          })
          return
        }
      })
      this.feedback[index][id] =
        this.feedback[index][id] || (value ? '' : get(_)('phrases.NameMustNotBeEmpty'))
    } else if (key == 'url') {
      this.feedback[index][id] =
        value.startsWith('http://') || value.startsWith('https://')
          ? ''
          : get(_)('phrases.URLMustBeginWithHttp')
    } else if (key == 'host' && value === '') {
      this.feedback[index][id] = get(_)('phrases.HostMustNotBeEmpty')
    } else if (key == 'apiKey' && value.length < 32) {
      this.feedback[index][id] = get(_)('phrases.APIKeyMustBeCountCharacters', {
        values: { count: 32 },
      })
    } else if (key == 'token' && value.length < 8) {
      this.feedback[index][id] = get(_)('phrases.TokenMustBeCountCharacters', {
        values: { count: 8 },
      })
    }

    if (this.app.customValidator) {
      const feedback = this.app.customValidator(id, value, index)
      if (feedback !== undefined) this.feedback[index][id] = feedback
    }

    return this.feedback[index][id]
  }
}
