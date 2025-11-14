class stdCache {
  private _value: string = ''
  wrapWrite (std: any): any {
    const write = std.write
    const thisObj = this
    // @ts-ignore
    return function (string, encoding, fileDescriptor) {
      thisObj._value += string
      write.apply(std, arguments)
    }
  }

  get content (): string {
    return this._value
  }
}

export interface StdCache {
  content: string
}

export function cacheStd (): StdCache {
  const cache: stdCache = new stdCache()
  process.stdout.write = cache.wrapWrite(process.stdout)
  process.stderr.write = cache.wrapWrite(process.stderr)
  return cache
}
