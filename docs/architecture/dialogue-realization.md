# Realizacja odpowiedzi dialogowej

`internal/dialogue` przekazuje `planner.ContentPlan` do tego samego `engine.TargetRealizer`, którego używa tłumaczenie. Dzięki temu wybór leksykalizacji, walidacja walencji, kompilowany `ConstructionPlan`, morfologia i linearyzacja pozostają wspólną ścieżką runtime.

Obecny most dopuszcza wyłącznie odpowiedź `statement` z celem `answer_yes_no` albo `answer_wh`, zawierającą jeden predykat/referent i opcjonalną jawną negację. Wynik zachowuje identyfikatory faktów, pochodzenie dowodów, proof refs, ślady target pipeline oraz ID planu, rodziny i wariantu konstrukcji.

Powierzchnia jest usuwana z wyniku, jeśli target pipeline zgłosi błąd, stratę pola albo brak pełnego śladu zatwierdzonej konstrukcji. Most nie wybiera jednej z konkurencyjnych powierzchni.

Pytanie, doprecyzowanie, odmowa i wyjaśnienie mają osobną siłę dialogową. Dopóki język nie ma zweryfikowanej konstrukcji dla danego aktu, wynik to jawne `unsupported`; nie ma fallbacku do zwykłego zdania ani ad hoc tekstu. Pakiety PL/EN obecnie nie dostarczają zatwierdzonych konstrukcji odpowiedzi dla tych aktów, więc ich mapowania i ręczna ocena pozostają otwarte.

Biblioteka nie jest podłączona do CLI dialogowego. `translate` nadal zgłasza brak skonfigurowanego runtime; owner-provided manual review jest **NOT RUN**.
