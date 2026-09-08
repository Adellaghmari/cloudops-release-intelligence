describe('Recruiter local workflow', () => {
  it('shows operations overview and a synthetic release decision surface', () => {
    cy.visit('/');
    cy.contains('CloudOps');
    cy.contains('Change. Risk. Impact. Recovery.');
    cy.contains('SYNTHETIC DEMO');
    cy.contains('System Status');
    cy.contains('a', 'Releases').click();
    cy.contains('SAFE_RELEASE');
    cy.contains('a', 'rel_northstar_payments_demo').click();
    cy.contains('Risk');
    cy.contains('Health');
    cy.contains('Potential impact');
    cy.contains('Policy');
    cy.contains('Rollback');
    cy.contains('Timeline');
    cy.contains('Release risk signal, not a probability of failure.');
    cy.contains('a', 'Replay').click();
    cy.contains('Release Replay');
    cy.contains('risk.score');
  });
});
