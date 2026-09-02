import { deepCopy, deepEqual, delay } from './util'
import type { Props as FaProps, IconSource } from './Fa.svelte'
import type { Config } from '../api/notifiarrConfig'

/**
 * App is a type that defines the structure of an application, instance or integration.
 * It is used to track the state of a form and its instances.
 * @param id - The id of the app. (StarrApps.Sonarr)
 * @param name - The name of the app. (Sonarr)
 * @param logo - The imported logo or Phosphor icon of the app. (../../assets/logos/sonarr.png)
 */
export type App<T> = {
  /** The id of the app. (StarrApps.Sonarr) */
  id: string
  /** The name of the app. (Sonarr) */
  name: string
  /** The imported logo or Phosphor icon of the app. (../../assets/logos/sonarr.png) */
  logo: string | IconSource
  /** Extra Fa props when `logo` is a Phosphor component. */
  iconProps?: Omit<FaProps, 'i' | 'children'>
  /** The disabled fields of the app. (['apiKey', 'username']) */
  disabled?: string[]
  /** The hidden fields of the app. (['deletes']) */
  hidden?: string[]
  /** The empty version of the form of the app. */
  empty?: T
  /** The environment prefix for this integration. */
  envPrefix: string

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
  validator?: (
    id: string,
    value: any,
    index: number,
    instances: T[],
  ) => string | undefined
}

/**
 * FormListTracker is a class that tracks multiple forms (across accordions generally).
 * it keeps track of the original list of instances, the form-bound list of instances,
 * the removed instances, and whether any instance is invalid.
 * @param instances - The form-bound list of instances in our tabs.
 * @param app - The app we're validating.
 */
export class FormListTracker<T> {
  /** The form-bound list of instances in our tabs. */
  public instances: T[]
  /** Count of deleted saved instances. Use .length for the Deleted badge. Indexes are not stable. */
  public removed: number[]
  /** Saved instances still in the form. Kept index-aligned with `instances` when deleting. */
  public readonly original: T[]
  /** Data about the app we're validating. */
  public readonly app: App<T>
  /** If any instance in the list fails validation the form is invalid. */
  public readonly invalid: boolean
  /** If the form has changed from the original values. */
  public readonly formChanged: boolean
  /** The active instance tab. */
  public active: number | undefined

  constructor(instances: T[], app: App<T>) {
    this.instances = $state(deepCopy(instances ?? []))
    this.original = $state(deepCopy(instances ?? []))
    this.app = app
    this.removed = $state([])
    this.active = $state(0)
    this.formChanged = $derived(
      this.removed.length > 0 || !deepEqual(this.instances, this.original),
    )
    // Must not write $state from validate(): Input derives feedback from it.
    this.invalid = $derived(this.instances.some((_, i) => !this.isValid(i)))
  }

  /** Add a new instance to the list. */
  public addInstance = () => {
    // Copy the empty template so each new instance is its own object.
    this.instances.push(deepCopy(this.app.empty!))
    this.active = this.instances.length - 1
  }

  /** Remove an instance from the list. */
  public delInstance = async (index: number) => {
    // Close the accordion.
    this.active = undefined
    // Wait for it to slide shut.
    await delay(400)
    // Drop the matching saved original so remaining rows keep the right Reset/Changed state.
    if (index < this.original.length) {
      this.original.splice(index, 1)
      this.removed.push(index)
    }
    // Remove the instance from the form (delete the accordion).
    this.instances.splice(index, 1)
    // Re-open a remaining instance. Never select index 0 on an empty list:
    // {#key flt.active} would remount children bound to undefined form and freeze the UI.
    this.active = this.instances.length
      ? Math.min(index, this.instances.length - 1)
      : undefined
  }

  /** Reset the form to the original values. Call this after a form has been submitted. */
  public resetAll = () => {
    this.instances = deepCopy(this.original)
    this.removed = []
  }

  /** Reset a single instance to the original values. Call this when reset button is clicked. */
  public resetForm = (index: number) => {
    this.instances[index] = deepCopy(this.original[index] ?? this.app.empty!)
  }

  /** Check if an instance is valid.
   * @param index - The index of the current instance the instances list. (0)
   */
  public isValid = (index: number): boolean => {
    const inst = this.instances[index]
    if (inst == null) return true
    return Object.keys(inst).every(
      k => !this.app.validator?.(this.app.id + '.' + k, inst[k as keyof T], index, this.instances),
    )
  }

  /** Standard form validator for an integrated instance (plex, sonarr, etc).
   * Pure: Input calls this from `$derived`, so it must not write `$state`.
   * @param id - The id of the form field. (anything.here.url)
   * @param value - The value of the form field. (http://localhost:8080)
   * @param index - The index of the current instance the instances list. (0)
   */
  public validate = (id: string, value: any, index: number): string | undefined => {
    return this.app.validator?.(id, value, index, this.instances)
  }
}
