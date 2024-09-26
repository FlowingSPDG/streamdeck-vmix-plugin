// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Fn = (...args: any[]) => void

export class EventListener<EventMap extends Record<string, Fn>> {
  private readonly callbacks = new Map<keyof EventMap, Set<Fn>>()

  add<K extends keyof EventMap>(event: K, callback: EventMap[K]): void {
    if (!this.callbacks.has(event)) {
      this.callbacks.set(event, new Set())
    }
    this.callbacks.get(event)!.add(callback)
  }

  remove<K extends keyof EventMap>(event: K, callback: EventMap[K]): void {
    this.callbacks.get(event)?.delete(callback)
  }

  dispatch<K extends keyof EventMap>(event: K, ...args: Parameters<EventMap[K]>): void {
    for (const callback of this.callbacks.get(event) ?? []) {
      callback(...args)
    }
  }
}
