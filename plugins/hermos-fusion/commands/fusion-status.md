---
description: Status eines einzelnen Fusion-Geräts
argument-hint: [geraetename-oder-id]
---

Status für: $ARGUMENTS

1. Ist kein Argument da, nach Gerät fragen. Ist es ein Name und keine GUID, über
   `list_devices` auflösen; bei mehreren Treffern zur Auswahl stellen.
2. `get_device` für die Stammdaten, `get_device_diagnostics` für die Lage.
3. Kompakt antworten, höchstens zehn Zeilen:
   - Name, Org-Unit, Ort
   - Online oder nicht, mit `lastSeenAt` in lokaler Zeit
   - Agent-Version, dazu ob ein Update ansteht oder erforderlich ist
   - Auffälligkeiten aus der Diagnose: Queue-Stau, fehlgeschlagene Befehle,
     hängende Transfers
4. Nichts verändern. Fällt eine Massnahme auf, vorschlagen statt ausführen.