import { Pipe, PipeTransform, inject } from '@angular/core';
import { DataSource } from '../core/models';
import {
  formatEvidenceAbsolute,
  formatEvidenceRelativeDay,
  formatEvidenceRelativeLabel,
} from '../core/format-evidence-time';
import { SyntheticDemoClock } from '../core/synthetic-demo-clock';

export type EvidenceTimeMode = 'absolute' | 'relative' | 'relativeDay';

@Pipe({
  name: 'evidenceTime',
  standalone: true,
  pure: false,
})
export class EvidenceTimePipe implements PipeTransform {
  private readonly demoClock = inject(SyntheticDemoClock);

  transform(
    iso: string | null | undefined,
    source: DataSource,
    mode: EvidenceTimeMode = 'absolute',
    reference = new Date(),
  ): string {
    if (!iso) {
      return '';
    }

    const instant =
      source === 'synthetic' ? this.demoClock.displayInstant(iso, reference) : new Date(iso);
    if (Number.isNaN(instant.getTime())) {
      return iso;
    }

    switch (mode) {
      case 'relativeDay':
        return formatEvidenceRelativeDay(instant, reference);
      case 'relative':
        return formatEvidenceRelativeLabel(instant, reference);
      default:
        return formatEvidenceAbsolute(instant);
    }
  }
}
