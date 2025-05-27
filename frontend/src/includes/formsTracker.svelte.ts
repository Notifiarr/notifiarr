import { deepCopy, deepEqual, delay } from './util'
import type { Props as FaProps } from './Fa.svelte'
import type { Config } from '../api/notifiarrConfig'
import type { IconDefinition } from '@fortawesome/sharp-duotone-light-svg-icons'

/**
 * App is a type that defines the structure of an application, instance or integration.
 * It is used to track the state of a form and its instances.
 * @param id - The id of the app. (StarrApps.Sonarr)
 * @param name - The name of the app. (Sonarr)
 * @param logo - The imported logo or FA icon of the app. (../../assets/logos/sonarr.png)
 */
export type App<T> = {
  /** The id of the app. (StarrApps.Sonarr) */
  id: string
  /** The name of the app. (Sonarr) */
  name: string
  /** The imported logo or FA icon of the app. (../../assets/logos/sonarr.png) */
  logo: string | IconDefinition
  /** If you provided an FA icon, you can use this to set the icon props. */
  iconProps?: Omit<FaProps, 'i'>
  /** The disabled fields of the app. (['apiKey', 'username']) */
  disabled?: string[]
  /** The hidden fields of the app. (['deletes']) */
  hidden?: string[]
  /** The empty version of the form of the app. */
  empty?: T
  /** The merge function of the app.
   * This is used when checking (testing) an instance.
   * The check function calls this to merge the instance with the original config.
   * @param index - The index of the instance.
   * @param form - The form of the instance.
   * @returns The merged application config.
   */
  merge: (index: number, form: T) => Config
  /** The custom validator of the app.
   * This optional function is used to add additional validation to an instance's form elements.
   * Return undefined if the validator does not apply to the validated field.
   * @param id - The id of the field.
   * @param value - The value of the field.
   * @param index - The index of the instance.
   * @returns The feedback of the field.
   */
  validator?: (id: string, value: any, index: number) => string
}

/**
 * FormListTracker is a class that tracks multiple forms (across accordions generally).
 * it keeps track of the original list of instances, the form-bound list of instances,
 * the removed instances, the invalid instances, and the feedback for the instances.
 * @param instances - The form-bound list of instances in our tabs.
 * @param app - The app we're validating.
 */
export class FormListTracker<T> {
  /** List of invalid instances. */
  private feedback: Record<number, Record<string, string>> = $state({})
  /** The form-bound list of instances in our tabs. */
  public instances: T[]
  /** List of removed instance indexes. Use .length to get the number of removed instances. */
  public removed: number[] = $state([])
  /** The original list of instances in our tabs. */
  public readonly original: T[]
  /** Data about the app we're validating. */
  public readonly app: App<T>
  /** If any instance in the list has non-empty feedback the form is invalid. */
  public readonly invalid: boolean = $derived(
    Object.values(this.feedback).some(v => Object.values(v).some(v => !!v)),
  )

  /** If the form has changed from the original values. */
  public readonly formChanged: boolean
  /** The active instance tab. */
  public active: number | undefined = $state(0)

  constructor(instances: T[], app: App<T>) {
    this.instances = $state(deepCopy(instances ?? []))
    this.original = $state(deepCopy(instances ?? []))
    this.app = app
    this.formChanged = $derived(!deepEqual(this.instances, this.original))
  }

  /** Add a new instance to the list. */
  public addInstance = () => {
    this.instances.push(this.app.empty!)
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
    // Record the removed index. Once. If it's not a new index.
    // This is the 'Deleted' counter.
    if (!this.removed.includes(index) && index < this.original.length)
      this.removed.push(index)
    // Reset the feedback for the instance.
    this.feedback = {}
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
    this.instances[index] = deepCopy(this.original[index] ?? this.app.empty!)
    Object.keys(this.instances[index] ?? {}).forEach(k => {
      this.validate(k, this.instances[index]?.[k as keyof T], index)
    })
  }

  /** Validate all instances. Call this after a form has been submitted to re-validate any backend changes. */
  private validateAll = () => {
    this.instances.forEach((m, i) => {
      Object.keys(m ?? {}).forEach(k => {
        this.validate(k, m?.[k as keyof T], i)
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
    if (!this.feedback[index]) this.feedback[index] = {}
    return (this.feedback[index][id] = this.app.validator?.(id, value, index) ?? '')
  }
}
