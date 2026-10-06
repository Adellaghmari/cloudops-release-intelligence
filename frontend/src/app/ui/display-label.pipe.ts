import { Pipe, PipeTransform } from '@angular/core';

const preservedTerms = new Map([
  ['api', 'API'],
  ['ci', 'CI'],
  ['id', 'ID'],
  ['n/a', 'N/A'],
  ['sha', 'SHA'],
]);

@Pipe({
  name: 'displayLabel',
  standalone: true,
})
export class DisplayLabelPipe implements PipeTransform {
  transform(value: string | null | undefined): string {
    if (!value) {
      return 'Unknown';
    }

    return value
      .trim()
      .replaceAll('_', ' ')
      .toLowerCase()
      .split(/\s+/)
      .map((term, index) => {
        const preserved = preservedTerms.get(term);
        if (preserved) {
          return preserved;
        }
        return index === 0 ? term.charAt(0).toUpperCase() + term.slice(1) : term;
      })
      .join(' ');
  }
}
