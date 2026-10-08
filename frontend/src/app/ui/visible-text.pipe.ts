import { Pipe, PipeTransform } from '@angular/core';
import { stripVisibleDashes } from '../core/visible-text';

@Pipe({
  name: 'visibleText',
  standalone: true,
})
export class VisibleTextPipe implements PipeTransform {
  transform(value: string | null | undefined): string {
    if (!value) {
      return '';
    }
    return stripVisibleDashes(value);
  }
}
