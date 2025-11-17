import { useState, useEffect } from "react";

const STORAGE_KEY = "b4_test_domains";
const DEFAULT_DOMAINS = [
  "youtube.com",
  "google.com",
  "facebook.com",
  "twitter.com",
  "instagram.com",
];

export function useTestDomains() {
  const [domains, setDomains] = useState<string[]>(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? (JSON.parse(stored) as string[]) : DEFAULT_DOMAINS;
  });

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(domains));
  }, [domains]);

  const addDomain = (domain: string) => {
    const cleaned = domain.trim().toLowerCase();
    if (cleaned && !domains.includes(cleaned)) {
      setDomains([...domains, cleaned]);
    }
  };

  const removeDomain = (domain: string) => {
    setDomains(domains.filter((d) => d !== domain));
  };

  const clearDomains = () => {
    setDomains([]);
  };

  const resetToDefaults = () => {
    setDomains(DEFAULT_DOMAINS);
  };

  return {
    domains,
    addDomain,
    removeDomain,
    clearDomains,
    resetToDefaults,
  };
}
