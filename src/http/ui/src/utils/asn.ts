import * as ipaddr from "ipaddr.js";

interface AsnInfo {
  id: string;
  name: string;
  prefixes: string[];
}

const ASN_STORAGE_KEY = "b4_asn_cache";

class AsnStorage {
  private cache: Record<string, AsnInfo> | null = null;
  private readonly lookupCache = new Map<string, AsnInfo | null>();
  private cacheTimeout: ReturnType<typeof setTimeout> | null = null;
  private readonly MAX_CACHE_SIZE = 10000;

  private loadCache(): Record<string, AsnInfo> {
    if (this.cache === null) {
      const data = localStorage.getItem(ASN_STORAGE_KEY);
      this.cache = data ? (JSON.parse(data) as Record<string, AsnInfo>) : {};
    }
    return this.cache;
  }

  private resetCacheTimeout() {
    if (this.cacheTimeout) clearTimeout(this.cacheTimeout);
    this.cacheTimeout = setTimeout(() => {
      this.cache = null;
      this.lookupCache.clear();
    }, 60000);
  }

  addAsn(asnId: string, name: string, prefixes: string[]) {
    const cache = this.loadCache();
    cache[asnId] = { id: asnId, name, prefixes };
    localStorage.setItem(ASN_STORAGE_KEY, JSON.stringify(cache));
    this.lookupCache.clear();
    this.resetCacheTimeout();
  }

  getAll(): Record<string, AsnInfo> {
    return { ...this.loadCache() };
  }

  findAsnForIp(ip: string): AsnInfo | null {
    const cleanIp = ip.split(":")[0].replaceAll(/[[\]]/g, "");

    const cached = this.lookupCache.get(cleanIp);
    if (cached !== undefined) {
      // LRU refresh: move to end
      this.lookupCache.delete(cleanIp);
      this.lookupCache.set(cleanIp, cached);
      this.resetCacheTimeout();
      return cached;
    }

    const cache = this.loadCache();

    let result: AsnInfo | null = null;
    outer: for (const asn of Object.values(cache)) {
      for (const prefix of asn.prefixes) {
        if (this.ipInCidr(cleanIp, prefix)) {
          result = asn;
          break outer;
        }
      }
    }

    // Enforce max size
    if (this.lookupCache.size >= this.MAX_CACHE_SIZE) {
      const firstKey = this.lookupCache.keys().next().value;
      if (firstKey) this.lookupCache.delete(firstKey);
    }

    this.lookupCache.set(cleanIp, result);
    this.resetCacheTimeout();

    return result;
  }

  private ipInCidr(ip: string, cidr: string): boolean {
    try {
      const addr = ipaddr.process(ip);
      const range = ipaddr.parseCIDR(cidr);
      return addr.match(range);
    } catch {
      return false;
    }
  }

  clear() {
    localStorage.removeItem(ASN_STORAGE_KEY);
    this.cache = null;
    this.lookupCache.clear();
  }
}

export const asnStorage = new AsnStorage();
