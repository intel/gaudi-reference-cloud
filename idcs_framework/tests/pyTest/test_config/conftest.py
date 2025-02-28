import datetime

import os
import pytest

from pyTest.test_config.input import report_portal_info

TAGS = "--tags"


def pytest_sessionstart(session):
    """
    Called after the Session object has been created and
    before performing collection and entering the run test loop.
    """
    print("\n**** Building Test session environment starting ****")
    print("\n**** Session environment is ready to runs tests ****")


def pytest_addoption(parser):
    print("\n*** PYTEST ADDING TAG FUNCTIONALITY FOR MAT ***")
    # Testbed is not used anymore. In future if used then uncomment below
    # file_path = os.path.join(os.path.dirname(__file__), "testbed.json")
    parser.addini('suite_type', help="MAT", default="MAT")
    parser.addoption(TAGS, action="store",
                     help="enter valid vmware related tags")

    # overwrite report portal config
    test_details = report_portal_info
    if "rp_endpoint" in test_details and \
            test_details["rp_endpoint"] != "":
        parser.addini("rp_endpoint", 'help', type="pathlist")

    if "rp_uuid" in test_details and \
            test_details["rp_uuid"] != "":
        parser.addini("rp_uuid", 'help', type="pathlist")
    parser.addini("rp_uuid", 'help', type="pathlist")
    parser.addini("rp_endpoint", 'help', type="pathlist")
    parser.addini("rp_launch", 'help', type="pathlist")
    parser.addini("rp_launch_description", 'help', type="pathlist")
    parser.addini("rp_launch_attributes", 'help', type="pathlist")


def pytest_collection_modifyitems(config, items):
    """
    Filter test cases based on tags.
     This hook will de-select the testcases
     from test items if given tag is not present.
    :param config: pytest config object
    :param items: testcase object list
    """
    print(f"\n*** TESTS WITHOUT TAG FILTERING: {items} ***")
    try:
        if not config.getoption(TAGS):
            print(f"\n*** --tags option not provided ***")
            return
    except Exception as ex:
        print(f"\n*** --tags option not provided. Running all tests ***")
        return

    with open(config.getoption(TAGS)) as f:
        marker_lines = f.read().split("\n")
        print(f"\n*** TAGs in Input file: {marker_lines} ***")

    all_remaining = []
    all_deselected = []
    run_last = None

    for marker_line in marker_lines:
        # if empty string found as tag then skip
        if not marker_line:
            continue
        marker_line.strip()
        marker_line.strip(".")
        all_markers = marker_line.split(".")
        deselected = []
        remaining = []
        for marker in all_markers:
            # collect all test with marker
            if marker:
                for item in items:
                    # logic to run a specific test case in end.
                    # Can ignore this for now
                    if item.name == "test_created_session_objects":
                        run_last = item

                    # checking if any of testcase's tag is
                    # matching input tag then put in remaining list
                    item_markers = [mark.name == marker for
                                    mark in item.iter_markers()]
                    if any(item_markers) and item not in remaining:
                        remaining.append(item)

                    # remove item from remaining if tag doesnt match
                    elif not any(item_markers) and item in remaining:
                        remaining.pop(remaining.index(item))
                        if item not in deselected:
                            deselected.append(item)
                    elif not any(item_markers):
                        if item not in deselected:
                            deselected.append(item)

                    # if item is present in both deselected
                    # and remaining then remove it from remaining
                    if item in deselected and item in remaining:
                        remaining.pop(remaining.index(item))
        all_remaining.extend(remaining)
        all_deselected.extend(deselected)

    # keep only unique test case items
    all_remaining = set(all_remaining)
    all_deselected = set(all_deselected)

    # if all_deselected has TCs of remaining somehow then remove again
    for remaining in all_remaining:
        if remaining in all_deselected:
            print("*** Removing TESTS from deselected: ", remaining)
            all_deselected.remove(remaining)
    print(
        f"\n***Selected tests: {all_remaining} "
        f"\n skipping: {all_deselected}"
    )
    if all_deselected:
        config.hook.pytest_deselected(items=all_deselected)
        items[:] = all_remaining

    # sort the testcases, else while running
    # in parallel it gives issues in test collection
    items.sort(key=sort_by_cls_name)

    if run_last in items:
        items.remove(run_last)
        items.append(run_last)


def pytest_configure(config):
    """
    Allows plugins and conftest files to perform initial configuration.
    This hook is called for every plugin and initial conftest
    file after command line options have been parsed.
    """
    print("\n**** Configure TEST environment ****")
    # overwrite report portal config
    test_details = report_portal_info
    import pdb
    pdb.set_trace()
    if "rp_endpoint" in test_details and \
            test_details["rp_endpoint"] != "":
        config._inicache["rp_endpoint"] = \
            test_details["rp_endpoint"]

    if "rp_uuid" in test_details and \
            test_details["rp_uuid"] != "":
        config._inicache["rp_uuid"] = \
            test_details["rp_uuid"]
        os.environ["rp_uuid"] = \
            test_details["rp_uuid"]

    config._inicache["rp_launch"] = \
        test_details["rp_launch"]

    if "launch_description" in test_details:
        config._inicache["rp_launch_description"] = \
            test_details["launch_description"]
    attributes = []
    attributes.append("Test Version:v1.0")
    attributes.append(
        "Launch Time:" + datetime.datetime.now().strftime(
            "%Y/%m/%dT%H.%M.%S"))
    launch_attributes = test_details["launch_attributes"]
    for keys in launch_attributes:
        attributes.append(keys + ":" + str(launch_attributes[keys]))
    print(attributes)
    config._inicache["rp_launch_description"] = \
        test_details["launch_description"]
    config._inicache["rp_launch"] = \
        test_details["rp_launch"]
    config._inicache["rp_uuid"] = \
        test_details["rp_uuid"]
    os.environ["rp_uuid"] = \
        test_details["rp_uuid"]
    config._inicache["rp_endpoint"] = \
        test_details["rp_endpoint"]
    config._inicache["rp_launch_attributes"] = attributes


def sort_by_cls_name(elem):
    return str(elem.cls) + str(elem.originalname)


def pytest_sessionfinish(session):
    """
    Called after whole test run finished, right before
    returning the exit status to the system.
    """
    print(f"......processids-> {os.getpid()} {os.getppid()}")
    print("\n**** Cleaning Test session ****")
    print("\n**** Session objects are destroyed successfully!! ****")
    print("*** getting  created objects from cache ****")


def pytest_unconfigure():
    """
    called
    """
    print("\n**** TEARDOWN Test environment ****")
