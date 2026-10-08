import { Injectable } from '@angular/core';
import { startOfLocalDay } from './format-evidence-time';

@Injectable({ providedIn: 'root' })
export class SyntheticDemoClock {
  private anchorDayStart: number | null = null;

  noteTimestamps(isoValues: (string | null | undefined)[]) {
    for (const iso of isoValues) {
      if (!iso) {
        continue;
      }
      const instant = new Date(iso);
      if (Number.isNaN(instant.getTime())) {
        continue;
      }
      const dayStart = startOfLocalDay(instant);
      if (this.anchorDayStart === null || dayStart > this.anchorDayStart) {
        this.anchorDayStart = dayStart;
      }
    }
  }

  displayInstant(iso: string, reference = new Date()): Date {
    const original = new Date(iso);
    if (Number.isNaN(original.getTime())) {
      return original;
    }
    if (this.anchorDayStart === null) {
      this.noteTimestamps([iso]);
    }
    const shift = startOfLocalDay(reference) - (this.anchorDayStart ?? startOfLocalDay(original));
    return new Date(original.getTime() + shift);
  }
}

export function collectSyntheticTimestamps(values: {
  created_at?: string;
  committed_at?: string;
  started_at?: string;
  completed_at?: string;
  assessed_at?: string;
  evaluated_at?: string;
  compared_at?: string;
  baseline_from?: string;
  baseline_to?: string;
  post_from?: string;
  post_to?: string;
  occurred_at?: string;
}[]): string[] {
  const out: string[] = [];
  for (const value of values) {
    for (const iso of Object.values(value)) {
      if (typeof iso === 'string' && iso) {
        out.push(iso);
      }
    }
  }
  return out;
}
