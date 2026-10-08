export function startOfLocalDay(date: Date): number {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
}

const timeFormatter = new Intl.DateTimeFormat('en-GB', {
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
});

const absoluteFormatter = new Intl.DateTimeFormat('en-GB', {
  day: 'numeric',
  month: 'short',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
});

export function formatEvidenceTimeOfDay(date: Date): string {
  return timeFormatter.format(date).replace(',', '');
}

export function formatEvidenceAbsolute(date: Date): string {
  const parts = absoluteFormatter.formatToParts(date);
  const day = parts.find((p) => p.type === 'day')?.value ?? '';
  const month = parts.find((p) => p.type === 'month')?.value ?? '';
  const year = parts.find((p) => p.type === 'year')?.value ?? '';
  const hour = parts.find((p) => p.type === 'hour')?.value ?? '';
  const minute = parts.find((p) => p.type === 'minute')?.value ?? '';
  return `${day} ${month} ${year} · ${hour}:${minute}`;
}

export function formatEvidenceRelativeLabel(date: Date, reference = new Date()): string {
  const time = formatEvidenceTimeOfDay(date);
  const dayDiff = Math.round((startOfLocalDay(reference) - startOfLocalDay(date)) / 86_400_000);

  if (dayDiff <= 0) {
    return `Today · ${time}`;
  }
  if (dayDiff === 1) {
    return `Yesterday · ${time}`;
  }
  return `${dayDiff} days ago · ${time}`;
}

export function formatEvidenceRelativeDay(date: Date, reference = new Date()): string {
  const dayDiff = Math.round((startOfLocalDay(reference) - startOfLocalDay(date)) / 86_400_000);

  if (dayDiff <= 0) {
    return 'Today';
  }
  if (dayDiff === 1) {
    return 'Yesterday';
  }
  return `${dayDiff} days ago`;
}
