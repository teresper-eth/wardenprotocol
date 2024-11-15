// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.25 <0.9.0;

struct ExecutionData {
    address caller;
    address to;
    uint256 chainId;
    bytes data;
    uint256 value;
}

enum Caller {
    Scheduler,
    AI
}

interface IExecution {
    /**
     * @dev Indicates if an order can be executed.
     * @return A boolean value indicating that the order can be executed.
     */
    function canExecute() external view returns (bool);

    /**
     * @dev Creates action for new sign request from stored order data.
     * If action created successfully then emit Executed event.
     * @param nonce The key account nonce.
     * @param gas gas parameter in eth transaction.
     * @param gasPrice gasPrice parameter in eth transaction.
     * @param maxPriorityFeePerGas maxPriorityFeePerGas parameter in eth transaction.
     * @param maxFeePerGas maxFeePerGas parameter in eth transaction.
     * @return A boolean value indicating parameter action was created.
     */
    function execute(
        uint256 nonce,
        uint256 gas,
        uint256 gasPrice,
        uint256 maxPriorityFeePerGas,
        uint256 maxFeePerGas
    )
        external
        returns (bool);

    /**
     * @dev Returns the list of intended callers.
     */
    function callers() external returns (Caller[] memory callersList);

    /**
     * @dev Receives data by AI service.
     * @param data Data from AI service.
     * @return success A boolean value indicating whenever data was processed correctly.
     */
    function setByAIService(bytes calldata data) external returns (bool success);

    /**
     * @dev Indicates if execution was executed already.
     */
    function isExecuted() external returns (bool);

    /**
     * @dev Returns data necessary for scheduler.
     * @return executionData Data that scheduler needs to call execute method.
     */
    function executionData() external returns (ExecutionData memory executionData);
}