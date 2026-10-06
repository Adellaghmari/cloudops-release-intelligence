describe('Recruiter local workflow', () => {
  it('moves from attention to evidence, services and semantic replay', () => {
    cy.visit('/');
    cy.contains('CloudOps');
    cy.contains('Change. Risk. Impact. Recovery.');
    cy.contains('Release decision support');
    cy.contains('Needs investigation');
    cy.contains('SYNTHETIC DEMO');
    cy.contains('a', 'Investigate release').click();
    cy.contains('Decision brief');
    cy.contains('No human release decision is recorded');

    cy.contains('a', 'Health').click();
    cy.location('hash').should('eq', '#health');
    cy.get('#health').should('be.visible');
    cy.contains('a', 'Impact').click();
    cy.location('hash').should('eq', '#impact');
    cy.get('#impact').should('be.visible');
    cy.contains('a', 'Policy').click();
    cy.location('hash').should('eq', '#policy');
    cy.contains('a', 'Rollback').click();
    cy.location('hash').should('eq', '#rollback');
    cy.contains('a', 'Timeline').click();
    cy.location('hash').should('eq', '#timeline');

    cy.contains('a', 'Compare release').first().click();
    cy.contains('Release Replay');
    cy.get('select').eq(1).select('rel_northstar_payments_demo');
    cy.contains('Comparison summary');
    cy.contains('Risk');
    cy.contains('Health');
    cy.contains('Recovery');
    cy.contains('Changed only').click();

    cy.contains('a', 'Services').click();
    cy.contains('Change impact catalog');
    cy.contains('a', 'Explore relationships').first().click();
    cy.contains('Related releases');

    cy.visit('/architecture');
    cy.contains('Verified architecture');
    cy.contains('Release Risk');
    cy.contains('Data contract');
  });
});

describe('Public CloudFront recruiter smoke', () => {
  const base = Cypress.env('PUBLIC_BASE_URL');
  const runPublic = Boolean(base);

  (runPublic ? it : it.skip)('loads public SPA and API-backed release surfaces', () => {
    cy.visit(base as string);
    cy.contains('CloudOps');
    cy.contains('LIVE PROJECT DATA');
    cy.contains('a', 'Releases').click();
    cy.contains('a', 'rel_northstar_payments_demo', { timeout: 20000 }).click();
    cy.contains('Risk', { timeout: 20000 });
    cy.contains('Health');
    cy.contains('SYNTHETIC DEMO');
  });
});

