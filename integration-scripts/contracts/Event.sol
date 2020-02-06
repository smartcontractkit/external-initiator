pragma solidity ^0.4.23;

contract Event {
  event SomeEvent(bytes32 data);

  function logEvent(bytes32 _data) public {
    emit SomeEvent(_data);
  }
}
