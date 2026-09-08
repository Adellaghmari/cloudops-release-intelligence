import { Routes } from '@angular/router';
import { Shell } from './layout/shell';
import { ArchitecturePage } from './pages/architecture.page';
import { OverviewPage } from './pages/overview.page';
import { ReleaseDetailPage } from './pages/release-detail.page';
import { ReleasesPage } from './pages/releases.page';
import { ServiceDetailPage } from './pages/service-detail.page';
import { ServicesPage } from './pages/services.page';

export const routes: Routes = [
  {
    path: '',
    component: Shell,
    children: [
      { path: '', component: OverviewPage },
      { path: 'services', component: ServicesPage },
      { path: 'services/:id', component: ServiceDetailPage },
      { path: 'releases', component: ReleasesPage },
      { path: 'releases/:id', component: ReleaseDetailPage },
      { path: 'architecture', component: ArchitecturePage },
    ],
  },
];
