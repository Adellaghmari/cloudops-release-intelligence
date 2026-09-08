import { HttpInterceptorFn } from '@angular/common/http';

export const requestIdInterceptor: HttpInterceptorFn = (req, next) => {
  if (req.headers.has('X-Request-ID')) {
    return next(req);
  }
  const id = `req_web_${crypto.randomUUID().replaceAll('-', '').slice(0, 20)}`;
  return next(req.clone({ setHeaders: { 'X-Request-ID': id } }));
};
