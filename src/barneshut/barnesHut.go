package barneshut

import (
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
)

const SOFTENING = 0.000000001

const THETA = 0.5

type Particle struct {
	// Assuming unit mass
	x, y, vx, vy, fx, fy float64
}

func NewParticle(x float64, y float64) *Particle {
	particle := new(Particle)
	particle.x, particle.y = x, y
	particle.vx, particle.vy, particle.fx, particle.fy = 0.0, 0.0, 0.0, 0.0
	return particle
}

type BarnesHutNode struct {
	centerX, centerY                     float64 // Used to divide the subquadrants.
	totalMass                            float64 // 1.0 if 1 particle else total mass of the children.
	comX, comY                           float64 // Center of Mass X & Y positions.
	leftX, rightX, topY, botY            float64 // Bounds for the quadrant.
	particle                             *Particle
	topLeft, botLeft, topRight, botRight *BarnesHutNode
}

/*
** Create Nodes for the Quadrants.
 */
func CreateNode(leftX float64, rightX float64, botY float64, topY float64, particle *Particle) *BarnesHutNode {
	node := new(BarnesHutNode)
	node.leftX, node.rightX, node.botY, node.topY = leftX, rightX, botY, topY
	node.centerX = leftX + (rightX-leftX)/2.0 // Required during further divisions of quadrant.
	node.centerY = botY + (topY-botY)/2.0
	node.particle = particle
	node.topLeft, node.topRight, node.botLeft, node.botRight = nil, nil, nil, nil
	node.comX = 0.0
	node.comY = 0.0
	return node
}

/*
** Inserts a Particle in the appropriate quadrant.
 */
func InsertParticle(node *BarnesHutNode, particle *Particle) {
	if node.particle == nil && node.topLeft == nil && node.topRight == nil && node.botLeft == nil && node.botRight == nil {
		// Set node particle for leaf node.
		node.particle = particle
	} else if node.particle != nil {
		// Node already contains a particle so subdivide and reassign particles...
		var avgX float64 = node.leftX + ((node.rightX - node.leftX) / 2.0) // avoids overflow due to addition of max vals.
		var avgY float64 = node.botY + ((node.topY - node.botY) / 2.0)
		// Create the quadrants.
		node.topLeft = CreateNode(node.leftX, avgX, avgY, node.topY, nil)
		node.botLeft = CreateNode(node.leftX, avgX, node.botY, avgY, nil)
		node.topRight = CreateNode(avgX, node.rightX, avgY, node.topY, nil)
		node.botRight = CreateNode(avgX, node.rightX, node.botY, avgY, nil)

		// Insert the existing particle to the appropriate quadrant
		var currentNodeParticle *Particle = node.particle
		node.particle = nil // We need to make this nil before recursion to avoid infinite recursion.
		InsertParticle(node, currentNodeParticle)
		// Insert the new particle in the appropriate quadrant.
		InsertParticle(node, particle)
	} else {
		// Node doesn't conatain a particle and is already subdivided.
		// Insert recursively into the right quadrant.
		var avgX float64 = node.leftX + ((node.rightX - node.leftX) / 2.0)
		var avgY float64 = node.botY + ((node.topY - node.botY) / 2.0)
		if particle.x < node.centerX && particle.y >= node.centerY {
			// Top Left Quadrant
			if node.topLeft == nil {
				node.topLeft = CreateNode(node.leftX, avgX, avgY, node.topY, nil)
			}
			InsertParticle(node.topLeft, particle)
		} else if particle.x < node.centerX && particle.y < node.centerY {
			// Bottom Left Quadrant
			if node.botLeft == nil {
				node.botLeft = CreateNode(node.leftX, avgX, node.botY, avgY, nil)
			}
			InsertParticle(node.botLeft, particle)
		} else if particle.x >= node.centerX && particle.y >= node.centerY {
			// Top Right Quadrant
			if node.topRight == nil {
				node.topRight = CreateNode(avgX, node.rightX, avgY, node.topY, nil)
			}
			InsertParticle(node.topRight, particle)
		} else {
			// Bottom Right Quadrant
			if node.botRight == nil {
				node.botRight = CreateNode(avgX, node.rightX, node.botY, avgY, nil)
			}
			InsertParticle(node.botRight, particle)
		}
	}
}

/*
** Calculates the Center of Mass for all the quadrants, and subquadrants.
** Assigns same values as the particle if quadrant conatains a particle.
 */
func CalcCenterOfMass(node *BarnesHutNode) {
	if node == nil {
		return
	}

	if node.topLeft == nil && node.topRight == nil && node.botLeft == nil && node.botRight == nil {
		// In leaf node the COM would be the same as the particle.
		if node.particle != nil {
			node.totalMass = 1.0
			node.comX = node.particle.x
			node.comY = node.particle.y
		}
	} else {
		var totalMass float64 = 0.0
		var comX float64 = 0.0
		var comY float64 = 0.0

		// Recursively calculate for each non-nil subquadrant.
		if node.topLeft != nil {
			CalcCenterOfMass(node.topLeft)
			if node.topLeft.totalMass > 0.0 {
				// Need this if check for unpruned tree
				totalMass += node.topLeft.totalMass
				comX += node.topLeft.comX * node.topLeft.totalMass
				comY += node.topLeft.comY * node.topLeft.totalMass
			}
		}

		if node.topRight != nil {
			CalcCenterOfMass(node.topRight)
			if node.topRight.totalMass > 0.0 {
				totalMass += node.topRight.totalMass
				comX += node.topRight.comX * node.topRight.totalMass
				comY += node.topRight.comY * node.topRight.totalMass
			}
		}

		if node.botLeft != nil {
			CalcCenterOfMass(node.botLeft)
			if node.botLeft.totalMass > 0.0 {
				totalMass += node.botLeft.totalMass
				comX += node.botLeft.comX * node.botLeft.totalMass
				comY += node.botLeft.comY * node.botLeft.totalMass
			}
		}

		if node.botRight != nil {
			CalcCenterOfMass(node.botRight)
			if node.botRight.totalMass > 0.0 {
				totalMass += node.botRight.totalMass
				comX += node.botRight.comX * node.botRight.totalMass
				comY += node.botRight.comY * node.botRight.totalMass
			}
		}

		// Calculate COM for the current node
		node.totalMass = totalMass
		if totalMass > 0.0 {
			// avoid 0 division error
			node.comX = comX / totalMass
			node.comY = comY / totalMass
		} else {
			node.comX = 0.0
			node.comY = 0.0
		}
	}
}

/*
** Calculates the force on a particle by a node or a particle in the node,
** Adds the force component to the force data member in the particle instance.
 */
func ForceByNode(particle *Particle, node *BarnesHutNode) {
	var dx float64 = node.comX - particle.x
	var dy float64 = node.comY - particle.y
	var distSqr float64 = dx*dx + dy*dy + SOFTENING
	var invDist float64 = 1.0 / math.Sqrt(distSqr)
	var invDist3 float64 = invDist * invDist * invDist
	particle.fx += dx * node.totalMass * invDist3
	particle.fy += node.totalMass * dy * invDist3
}

/*
** Calculates the net forces on a particle and stores in fx, fy data members.
 */
func ForceCalculation(particle *Particle, node *BarnesHutNode) {
	if node == nil {
		return
	}

	if node.particle == particle {
		// Same particle
		return
	}

	var dx float64 = particle.x - node.comX
	var dy float64 = particle.y - node.comY
	var distSqr float64 = dx*dx + dy*dy + SOFTENING
	var D float64 = math.Sqrt(distSqr)
	var S float64 = node.rightX - node.leftX // Width/Size of the quadrant.
	var sByD float64 = S / D

	if sByD < THETA || node.particle != nil {
		// Either Particle in node so leaf node
		// Or s/d is less than theta, so use COM.
		ForceByNode(particle, node)
	} else {
		ForceCalculation(particle, node.topLeft)
		ForceCalculation(particle, node.topRight)
		ForceCalculation(particle, node.botLeft)
		ForceCalculation(particle, node.botRight)
	}
}

/*
** calc and store the new velocity of the particle
 */
func CalcVelocity(particle *Particle, root *BarnesHutNode, dt float64) {
	ForceCalculation(particle, root)
	particle.vx += dt * particle.fx
	particle.vy += dt * particle.fy
}

/*
** Updates the new positions of the particles.
** And inserts the particle in the new Tree for recreating quadrants.
 */
func CalcNewPositions(root *BarnesHutNode, dt float64) {
	if root == nil {
		return
	}

	if root.particle != nil {
		root.particle.x += root.particle.vx * dt
		root.particle.y += root.particle.vy * dt
		// Reset forces for next iteration
		root.particle.fx, root.particle.fy = 0.0, 0.0
		// InsertParticle(newRoot, root.particle) // INsert in new tree.
	} else {
		CalcNewPositions(root.topLeft, dt)
		CalcNewPositions(root.topRight, dt)
		CalcNewPositions(root.botLeft, dt)
		CalcNewPositions(root.botRight, dt)
	}
}

/*
** Calculate and store the valocities of all the particles.
 */
func CalcVelocityForAll(node *BarnesHutNode, root *BarnesHutNode, dt float64) {
	if node == nil {
		return
	}

	if node.particle != nil {
		CalcVelocity(node.particle, root, dt)
	}
	CalcVelocityForAll(node.topLeft, root, dt)
	CalcVelocityForAll(node.topRight, root, dt)
	CalcVelocityForAll(node.botLeft, root, dt)
	CalcVelocityForAll(node.botRight, root, dt)
}

/*
** Recreates the new tree with the new position of the particles.
 */
func RecreateWithNewPos(root *BarnesHutNode, newRoot *BarnesHutNode) {
	if root == nil {
		return
	}

	if root.particle != nil {
		InsertParticle(newRoot, root.particle)
	} else {
		RecreateWithNewPos(root.topLeft, newRoot)
		RecreateWithNewPos(root.topRight, newRoot)
		RecreateWithNewPos(root.botLeft, newRoot)
		RecreateWithNewPos(root.botRight, newRoot)
	}
}

/*
** Print/Debug Utils
 */

func PrintBarnesHutTree(root *BarnesHutNode) {
	if root == nil {
		return
	}

	fmt.Printf("X: %f, Y: %f\n", root.comX, root.comY)

	PrintBarnesHutTree(root.topLeft)
	PrintBarnesHutTree(root.topRight)
	PrintBarnesHutTree(root.botLeft)
	PrintBarnesHutTree(root.botRight)
}

func PrintBarnesHutTreeParticle(root *BarnesHutNode) {
	if root == nil {
		return
	}
	if root.particle != nil {
		fmt.Printf("X: %f, Y: %f\n", root.particle.x, root.particle.y)
	}
	PrintBarnesHutTreeParticle(root.topLeft)
	PrintBarnesHutTreeParticle(root.topRight)
	PrintBarnesHutTreeParticle(root.botLeft)
	PrintBarnesHutTreeParticle(root.botRight)
}

/*
** Print the input and output positions to the .dat file
 */
func FprintDataFile(file *os.File, root *BarnesHutNode) {
	if root == nil {
		return
	}
	if root.particle != nil {
		fmt.Fprintf(file, "%f %f\n", root.particle.x, root.particle.y)
	}
	FprintDataFile(file, root.topLeft)
	FprintDataFile(file, root.topRight)
	FprintDataFile(file, root.botLeft)
	FprintDataFile(file, root.botRight)
}

/************* DEQUEU **************/

type Task struct {
	Node *BarnesHutNode
}

// Using Linked LIst implementation of Deque.
type node struct {
	task *Task
	next *node
	prev *node
}

type Deque struct {
	mu   sync.Mutex
	head *node
	tail *node
	size int32
}

func NewDeque() *Deque {
	return &Deque{}
}

func (d *Deque) PushFront(t Task) {
	d.mu.Lock()
	defer d.mu.Unlock()

	newNode := &node{task: &t}
	if d.head == nil {
		d.head = newNode
		d.tail = newNode
	} else {
		newNode.next = d.head
		d.head.prev = newNode
		d.head = newNode
	}
	d.size++
}

func (d *Deque) PushBack(t Task) {
	d.mu.Lock()
	defer d.mu.Unlock()

	newNode := &node{task: &t}
	if d.tail == nil {
		d.head = newNode
		d.tail = newNode
	} else {
		newNode.prev = d.tail
		d.tail.next = newNode
		d.tail = newNode
	}
	d.size++
}

func (d *Deque) PopFront() (Task, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.head == nil {
		return Task{}, false
	}
	t := *d.head.task
	d.head = d.head.next
	if d.head != nil {
		d.head.prev = nil
	} else {
		d.tail = nil
	}
	d.size--
	return t, true
}

func (d *Deque) PopBack() (Task, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.tail == nil {
		return Task{}, false
	}
	t := *d.tail.task
	d.tail = d.tail.prev
	if d.tail != nil {
		d.tail.next = nil
	} else {
		d.head = nil
	}
	d.size--
	return t, true
}

func (d *Deque) Len() int32 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.size
}

/******************** END OF DEQUE ************************/

/**** SUPERSTEP FUNCTIONS ******/

func CalcCenterOfMassParallel(node *BarnesHutNode, activeThreads *int32, numThreads int) {
	if node == nil {
		return
	}

	if node.topLeft == nil && node.topRight == nil && node.botLeft == nil && node.botRight == nil {
		// In leaf node the COM would be the same as the particle.
		if node.particle != nil {
			node.totalMass = 1.0
			node.comX = node.particle.x
			node.comY = node.particle.y
		}
	} else {
		var totalMass float64 = 0.0
		var comX float64 = 0.0
		var comY float64 = 0.0

		// Recursively calculate for each non-nil subquadrant.
		var wgChildren sync.WaitGroup
		if node.topLeft != nil {
			if atomic.LoadInt32(activeThreads) < int32(numThreads) {
				atomic.AddInt32(activeThreads, 1)
				wgChildren.Add(1)
				go func() {
					defer wgChildren.Done()
					defer atomic.AddInt32(activeThreads, -1)
					CalcCenterOfMassParallel(node.topLeft, activeThreads, numThreads)
				}()
			} else {
				CalcCenterOfMassParallel(node.topLeft, activeThreads, numThreads)
			}
		}
		if node.topRight != nil {
			if atomic.LoadInt32(activeThreads) < int32(numThreads) {
				atomic.AddInt32(activeThreads, 1)
				wgChildren.Add(1)
				go func() {
					defer wgChildren.Done()
					defer atomic.AddInt32(activeThreads, -1)
					CalcCenterOfMassParallel(node.topRight, activeThreads, numThreads)
				}()
			} else {
				CalcCenterOfMassParallel(node.topRight, activeThreads, numThreads)
			}
		}
		if node.botLeft != nil {
			if atomic.LoadInt32(activeThreads) < int32(numThreads) {
				atomic.AddInt32(activeThreads, 1)
				wgChildren.Add(1)
				go func() {
					defer wgChildren.Done()
					defer atomic.AddInt32(activeThreads, -1)
					CalcCenterOfMassParallel(node.botLeft, activeThreads, numThreads)
				}()
			} else {
				CalcCenterOfMassParallel(node.botLeft, activeThreads, numThreads)
			}
		}
		if node.botRight != nil {
			if atomic.LoadInt32(activeThreads) < int32(numThreads) {
				atomic.AddInt32(activeThreads, 1)
				wgChildren.Add(1)
				go func() {
					defer wgChildren.Done()
					defer atomic.AddInt32(activeThreads, -1)
					CalcCenterOfMassParallel(node.botRight, activeThreads, numThreads)
				}()
			} else {
				CalcCenterOfMassParallel(node.botRight, activeThreads, numThreads)
			}
		}
		wgChildren.Wait()

		if node.topLeft.totalMass > 0.0 {
			// Need this if check for unpruned tree
			totalMass += node.topLeft.totalMass
			comX += node.topLeft.comX * node.topLeft.totalMass
			comY += node.topLeft.comY * node.topLeft.totalMass
		}
		if node.topRight.totalMass > 0.0 {
			totalMass += node.topRight.totalMass
			comX += node.topRight.comX * node.topRight.totalMass
			comY += node.topRight.comY * node.topRight.totalMass
		}
		if node.botLeft.totalMass > 0.0 {
			totalMass += node.botLeft.totalMass
			comX += node.botLeft.comX * node.botLeft.totalMass
			comY += node.botLeft.comY * node.botLeft.totalMass
		}
		if node.botRight.totalMass > 0.0 {
			totalMass += node.botRight.totalMass
			comX += node.botRight.comX * node.botRight.totalMass
			comY += node.botRight.comY * node.botRight.totalMass
		}

		// Calculate COM for the current node
		node.totalMass = totalMass
		if totalMass > 0.0 {
			// avoid 0 division error
			node.comX = comX / totalMass
			node.comY = comY / totalMass
		} else {
			node.comX = 0.0
			node.comY = 0.0
		}
	}
}

func RunSimulation(root *BarnesHutNode, numThreads int, dt float64, nParticles int) {
	// Synchronization primitives
	var wg sync.WaitGroup
	var activeThreads int32 = 1
	var velocityTasksProcessed int32 = 0
	var positionTasksProcessed int32 = 0

	// Ensure center of mass is calculated first
	CalcCenterOfMassParallel(root, &activeThreads, numThreads)

	// Initialize work-stealing deques
	deques := make([]*Deque, numThreads)
	for i := 0; i < numThreads; i++ {
		deques[i] = NewDeque()
	}

	// Velocity Calculation Phase
	deques[0].PushFront(Task{Node: root})
	wg.Add(numThreads)
	for t := 0; t < numThreads; t++ {
		go func(threadNum int) {
			defer wg.Done()
			calcVelocityWorker(root, dt, threadNum, deques, numThreads, &velocityTasksProcessed, nParticles)
		}(t)
	}
	wg.Wait()

	// Reset deques and prepare for position update
	for i := 0; i < numThreads; i++ {
		deques[i] = NewDeque()
	}
	deques[0].PushFront(Task{Node: root})

	// Position Update Phase
	wg.Add(numThreads)
	for t := 0; t < numThreads; t++ {
		go func(threadNum int) {
			defer wg.Done()
			updatePositionWorker(dt, threadNum, deques, numThreads, &positionTasksProcessed, nParticles)
		}(t)
	}
	wg.Wait()
}

func calcVelocityWorker(root *BarnesHutNode, dt float64, threadNum int, deques []*Deque, numThreads int, tasksProcessed *int32, nParticles int) {
	for {
		task, found := deques[threadNum].PopFront()
		if !found {
			// Unable to find task in its own queue so stealing now.
			// WORK STEALING
			task = stealTask(threadNum, deques, numThreads)
			if task.Node == nil {
				// Keep checking until all particles processed.
				if atomic.LoadInt32(tasksProcessed) >= int32(nParticles) {
					return
				}
				continue
			}
		}
		processVelocitySubtree(root, task.Node, dt, threadNum, deques, tasksProcessed)
	}
}

func updatePositionWorker(dt float64, threadNum int, deques []*Deque, numThreads int, tasksProcessed *int32, nParticles int) {
	for {
		task, found := deques[threadNum].PopFront()
		if !found {
			// Unable to find task in its own queue so stealing now.
			// WORK STEALING
			task = stealTask(threadNum, deques, numThreads)
			if task.Node == nil {
				// Keep checking until all particles processed.
				if atomic.LoadInt32(tasksProcessed) >= int32(nParticles) {
					return
				}
				continue
			}
		}
		processPositionSubtree(dt, task.Node, threadNum, deques, tasksProcessed)
	}
}

// WORK STEALING
func stealTask(threadNum int, deques []*Deque, numThreads int) Task {
	for i := 0; i < numThreads; i++ {
		victim := (threadNum + i) % numThreads
		if victim == threadNum {
			continue
		}
		stolenTask, foundThisTime := deques[victim].PopBack()
		if foundThisTime {
			return stolenTask
		}
	}
	return Task{Node: nil}
}

func processVelocitySubtree(root, node *BarnesHutNode, dt float64, threadNum int, deques []*Deque, tasksProcessed *int32) {
	if node == nil {
		return
	}

	// Add child nodes as tasks to deque
	if node.topLeft != nil {
		deques[threadNum].PushFront(Task{node.topLeft})
	}
	if node.topRight != nil {
		deques[threadNum].PushFront(Task{node.topRight})
	}
	if node.botLeft != nil {
		deques[threadNum].PushFront(Task{node.botLeft})
	}
	if node.botRight != nil {
		deques[threadNum].PushFront(Task{node.botRight})
	}

	if node.particle != nil {
		// Process particle if it exists
		CalcVelocity(node.particle, root, dt)
		atomic.AddInt32(tasksProcessed, 1)
	}
}

func processPositionSubtree(dt float64, node *BarnesHutNode, threadNum int, deques []*Deque, tasksProcessed *int32) {
	if node == nil {
		return
	}

	// Add child nodes as tasks to deque
	if node.topLeft != nil {
		deques[threadNum].PushFront(Task{node.topLeft})
	}
	if node.topRight != nil {
		deques[threadNum].PushFront(Task{node.topRight})
	}
	if node.botLeft != nil {
		deques[threadNum].PushFront(Task{node.botLeft})
	}
	if node.botRight != nil {
		deques[threadNum].PushFront(Task{node.botRight})
	}

	// Update particle position if it exists
	if node.particle != nil {
		node.particle.x += node.particle.vx * dt
		node.particle.y += node.particle.vy * dt
		node.particle.fx, node.particle.fy = 0.0, 0.0
		atomic.AddInt32(tasksProcessed, 1)
	}
}
