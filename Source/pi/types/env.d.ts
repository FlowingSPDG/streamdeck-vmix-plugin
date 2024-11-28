interface Window {
  connectElgatoStreamDeckSocket: (
    inPort: number,
    inUUID: string,
    inRegisterEvent: string,
    inInfo: string,
    inActionInfo: string,
  ) => void
}
