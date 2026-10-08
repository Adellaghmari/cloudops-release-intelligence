import { Pipe, PipeTransform } from '@angular/core';
import { formatEvidenceRelativeDay } from '../core/format-evidence-time';

@Pipe({
  name: 'relativeDay',
  standalone: true,
  pure: false,
})
export class RelativeDayPipe implements PipeTransform {
  transform(value: string | null | undefined, reference = new Date()): string {
    if (!value) {
      return '';
    }

    const recorded = new Date(value);
    if (Number.isNaN(recorded.getTime())) {
      return value;
    }

    return formatEvidenceRelativeDay(recorded, reference);
  }
}
