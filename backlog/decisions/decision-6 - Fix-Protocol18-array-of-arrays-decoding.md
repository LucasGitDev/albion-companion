---
id: decision-6
title: Fix Protocol18 array-of-arrays decoding
date: '2026-10-31T19:20:00Z'
status: accepted
---
## Context
The fork's deserializeNestedArray (type 0x40) read a string, so PartyJoined[8] (GUID list) decoded wrong.
## Decision
Match SA DeserializeArrayInArray: count, then a full typed value per element ([]any). Covered by the engine end-to-end test. Worth back-porting to the albiondata-client fork.
