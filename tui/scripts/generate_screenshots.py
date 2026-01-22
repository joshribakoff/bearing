#!/usr/bin/env python3
"""Generate visual regression screenshots for all scenarios."""
import asyncio
import shutil
from pathlib import Path

# Add parent to path for imports
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from tests.mock_data import (
    create_normal_workspace,
    create_empty_workspace,
    create_overflow_workspace,
    create_long_names_workspace,
    create_single_workspace,
)
from bearing_tui.app import BearingApp


async def generate_all():
    """Generate screenshots for all scenarios."""
    out = Path(__file__).parent.parent / "screenshots"
    out.mkdir(exist_ok=True)

    workspaces_to_cleanup = []

    try:
        # === NORMAL SCENARIO ===
        print("Generating normal scenario...")
        ws = create_normal_workspace()
        workspaces_to_cleanup.append(ws)

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "normal-01-initial.svg"))

            await pilot.press("j")
            await pilot.pause()
            app.save_screenshot(str(out / "normal-02-project-selected.svg"))

            await pilot.press("1")
            await pilot.press("j")
            await pilot.pause()
            app.save_screenshot(str(out / "normal-03-worktree-selected.svg"))

            await pilot.press("0")
            await pilot.pause()
            app.save_screenshot(str(out / "normal-04-panel0-unfocused-highlight.svg"))

        # === EMPTY SCENARIO ===
        print("Generating empty scenario...")
        ws = create_empty_workspace()
        workspaces_to_cleanup.append(ws)

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "empty-01-no-projects.svg"))

        # === OVERFLOW SCENARIO ===
        print("Generating overflow scenario...")
        ws = create_overflow_workspace()
        workspaces_to_cleanup.append(ws)

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "overflow-01-initial.svg"))

            for _ in range(15):
                await pilot.press("j")
            await pilot.pause()
            app.save_screenshot(str(out / "overflow-02-scrolled.svg"))

            await pilot.press("1")
            await pilot.pause()
            app.save_screenshot(str(out / "overflow-03-many-worktrees.svg"))

        # === LONG NAMES SCENARIO ===
        print("Generating long names scenario...")
        ws = create_long_names_workspace()
        workspaces_to_cleanup.append(ws)

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "longnames-01-truncation.svg"))

            await pilot.press("1")
            await pilot.press("j")
            await pilot.pause()
            app.save_screenshot(str(out / "longnames-02-details.svg"))

        # === SINGLE ITEM SCENARIO ===
        print("Generating single item scenario...")
        ws = create_single_workspace()
        workspaces_to_cleanup.append(ws)

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "single-01-minimal.svg"))

        # === RESIZE SCENARIOS ===
        print("Generating resize scenarios...")
        ws = create_normal_workspace()
        workspaces_to_cleanup.append(ws)

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(60, 25)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "resize-01-narrow.svg"))

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(100, 15)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "resize-02-short.svg"))

        app = BearingApp(workspace=ws)
        async with app.run_test(size=(50, 12)) as pilot:
            await pilot.pause()
            app.save_screenshot(str(out / "resize-03-tiny.svg"))

        print(f"✓ Generated {len(list(out.glob('*.svg')))} screenshots in {out}/")

    finally:
        # Cleanup temp workspaces
        for ws in workspaces_to_cleanup:
            shutil.rmtree(ws, ignore_errors=True)


if __name__ == "__main__":
    asyncio.run(generate_all())
