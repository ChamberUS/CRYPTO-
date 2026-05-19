# ADR 0001 — Branding externo IAOS, interno BYX

## Contexto
O projeto usa BYX no código (paths internos, denom, módulos) e IAOS na experiência/branding.
A UX e integrações externas devem apresentar IAOS, mantendo compatibilidade interna.

## Decisão
- Externo (UX / docs / endpoints públicos): IAOS
- Interno (código/paths/denom): BYX, por enquanto
- Criar alias REST /iaos/* apontando para /byx/*

## Consequências
- Sem quebra de compatibilidade
- UX e integrações passam a consumir IAOS desde já
