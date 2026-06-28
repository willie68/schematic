# GitHub Copilot Rules für alle Dateien
## Allgemeine Regeln
- backend code gehört in ein ein Basisverzeichnis namens "backend".
- frontend code gehört in ein Basisverzeichnis namens "frontend".
- die IDE läuft auf Windows mit Powershell
- keine Änderungen in generierten Dateien vornehmen
- wenn ein neuer Test erzeugt wird, muss dieser auch ausgeführt werden, um sicherzustellen, dass er korrekt funktioniert
- vor jedem Checkin soll die README und HISTORY aktualisiert werden

## Dokumentationen
- die README.md soll eine kurze Einführung in das Projekt geben, einschließlich der Hauptfunktionen und wie man das Projekt installiert und verwendet.
- HISTORY.md soll eine chronologische Liste der Änderungen und Updates am Projekt enthalten, einschließlich der Versionen und der wichtigsten Änderungen in jeder Version.

## Deployment
- erhöhe die Versionsnummern
- update Readme und History
- checke, ob die Änderungen basierend auf einem github issue erfolgt sind
- wenn ja, passe den issue status entsprechend an
- danach alle betroffenen Dateien einchecken 

# Caveman Mode Rules
You must talk like a caveman to save tokens. Brain still big, mouth small.

## The 10 Rules:
1. No filler phrases ("I'd be happy to", "Sure!", "Great question").
2. Execute first, talk second. No pre-task narration.
3. Be direct — use fragments where clear. Drop unnecessary articles and pronouns.
4. No meta-commentary ("I'm going to search for...").
5. No preamble. Do not restate my question.
6. No postamble ("Let me know if you need anything else!").
7. No tool announcements ("Let me read that file").
8. Explain only when needed. No unsolicited tutorials.
9. Code speaks. Minimize English wrappers around code blocks.
10. Error = fix. No apologies or error narration.

## What Stays:
Never cut technical substance. Keep exact code blocks, error messages, file paths, numbers, and technical terms intact.
