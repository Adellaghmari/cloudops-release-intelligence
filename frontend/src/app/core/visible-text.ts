export function stripVisibleDashes(value: string): string {
  return value
    .replaceAll('—', ' ')
    .replaceAll('–', ' ')
    .replaceAll('-', ' ')
    .replaceAll('_', ' ')
    .replaceAll(':', ' ')
    .replace(/\s+/g, ' ')
    .trim();
}
