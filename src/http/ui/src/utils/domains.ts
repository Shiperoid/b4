import { useMemo } from "react";
import { ParsedLog, SortColumn } from "../components/organisms/domains/Table";
import { SortDirection } from "../components/atoms/common/SortableTableCell";

// Parse log line from string
export function parseLogLine(line: string): ParsedLog | null {
  // Example: 2025/10/13 22:41:12.466126 [INFO] SNI TCP: assets.alicdn.com 192.168.1.100:38894 -> 92.123.206.67:443
  const regex =
    /^(\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2}\.\d+)\s+\[INFO\]\s+SNI\s+(TCP|UDP)(?:\s+TARGET)?:\s+(\S+)\s+(\S+)\s+->\s+(\S+)$/;
  const match = line.match(regex);

  if (!match) return null;

  const [, timestamp, protocol, domain, source, destination] = match;
  const isTarget = line.includes("TARGET");

  return {
    timestamp,
    protocol: protocol as "TCP" | "UDP",
    isTarget,
    domain,
    source,
    destination,
    raw: line,
  };
}

// Generate domain variants from most specific to least specific
export function generateDomainVariants(domain: string): string[] {
  const parts = domain.split(".");
  const variants: string[] = [];

  // Generate from full domain to TLD+1 (e.g., example.com)
  for (let i = 0; i < parts.length - 1; i++) {
    variants.push(parts.slice(i).join("."));
  }

  return variants;
}

// Hook to parse logs
export function useParsedLogs(lines: string[]): ParsedLog[] {
  return useMemo(() => {
    return lines
      .map(parseLogLine)
      .filter((log): log is ParsedLog => log !== null);
  }, [lines]);
}

// Hook to filter logs
export function useFilteredLogs(
  parsedLogs: ParsedLog[],
  filter: string
): ParsedLog[] {
  return useMemo(() => {
    const f = filter.trim().toLowerCase();
    const filters = f
      .split("+")
      .map((s) => s.trim())
      .filter((s) => s.length > 0);

    if (filters.length === 0) {
      return parsedLogs;
    }

    // Group filters by field
    const fieldFilters: Record<string, string[]> = {};
    const globalFilters: string[] = [];

    filters.forEach((filterTerm) => {
      const colonIndex = filterTerm.indexOf(":");

      if (colonIndex > 0) {
        const field = filterTerm.substring(0, colonIndex);
        const value = filterTerm.substring(colonIndex + 1);

        if (!fieldFilters[field]) {
          fieldFilters[field] = [];
        }
        fieldFilters[field].push(value);
      } else {
        globalFilters.push(filterTerm);
      }
    });

    return parsedLogs.filter((log) => {
      // Check field-specific filters (OR within field, AND across fields)
      for (const [field, values] of Object.entries(fieldFilters)) {
        const fieldValue =
          log[field as keyof typeof log]?.toString().toLowerCase() || "";
        const matches = values.some((value) => fieldValue.includes(value));
        if (!matches) return false;
      }

      // Check global filters (must match at least one field)
      for (const filterTerm of globalFilters) {
        const matches =
          log.domain.toLowerCase().includes(filterTerm) ||
          log.source.toLowerCase().includes(filterTerm) ||
          log.protocol.toLowerCase().includes(filterTerm) ||
          log.destination.toLowerCase().includes(filterTerm);
        if (!matches) return false;
      }

      return true;
    });
  }, [parsedLogs, filter]);
}

// Hook to sort logs
export function useSortedLogs(
  filteredLogs: ParsedLog[],
  sortColumn: SortColumn | null,
  sortDirection: SortDirection
): ParsedLog[] {
  return useMemo(() => {
    if (!sortColumn || !sortDirection) {
      return filteredLogs;
    }

    const sorted = [...filteredLogs].sort((a, b) => {
      let aValue: any = a[sortColumn];
      let bValue: any = b[sortColumn];

      // Handle different data types
      if (sortColumn === "timestamp") {
        aValue = new Date(aValue.replace(/\//g, "-")).getTime();
        bValue = new Date(bValue.replace(/\//g, "-")).getTime();
      } else if (sortColumn === "isTarget") {
        aValue = aValue ? 1 : 0;
        bValue = bValue ? 1 : 0;
      } else if (typeof aValue === "string") {
        aValue = aValue.toLowerCase();
        bValue = bValue.toLowerCase();
      }

      if (aValue < bValue) {
        return sortDirection === "asc" ? -1 : 1;
      }
      if (aValue > bValue) {
        return sortDirection === "asc" ? 1 : -1;
      }
      return 0;
    });

    return sorted;
  }, [filteredLogs, sortColumn, sortDirection]);
}

// Local storage utilities
export const STORAGE_KEY = "b4_domains_lines";
export const MAX_STORED_LINES = 1000;

export function loadPersistedLines(): string[] {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      return Array.isArray(parsed) ? parsed : [];
    }
  } catch (e) {
    console.error("Failed to load persisted domains:", e);
  }
  return [];
}

export function persistLines(lines: string[]): void {
  try {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify(lines.slice(-MAX_STORED_LINES))
    );
  } catch (e) {
    console.error("Failed to persist domains:", e);
  }
}

export function clearPersistedLines(): void {
  localStorage.removeItem(STORAGE_KEY);
}
